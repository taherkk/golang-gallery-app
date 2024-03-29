package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/taherk/galleryapp/rand"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID     int
	UserID int
	// token is only set when a PasswordReset is being created
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	// how many bytes to use when generating each password reset token. If this vale is
	// not set or is less than the MinBytesPerToken const it will be ignored and
	// MinBytesPerToken will be used.
	BytesPerToken int
	// password expiration duration. Defaults to DefautlResetDuration
	Duration time.Duration
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {

	// verify we have a valid email and get user's id
	email = strings.ToLower(email)

	var userID int
	row := service.DB.QueryRow(`SELECT id from users WHERE email = $1;`, email)
	err := row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("models.passwordResetService.Create: %w", err)
	}

	// Build the password reset
	bytesPerToken := service.BytesPerToken
	if bytesPerToken < MinSessionTokenBytes {
		bytesPerToken = MinSessionTokenBytes
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("models.passwordResetService.Create: %w", err)
	}

	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	// insert the password reset into the db
	row = service.DB.QueryRow(`
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
		UPDATE
		SET token_hash = $2, expires_at = $3
		RETURNING id;`,
		pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt,
	)

	err = row.Scan(&pwReset.ID)

	if err != nil {
		return nil, fmt.Errorf("models.passwordResetService.Create: %w", err)
	}

	return &pwReset, nil
}

func (service *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := service.hash(token)
	var user User
	var pwReset PasswordReset

	row := service.DB.QueryRow(`
		SELECT password_resets.id,
			password_resets.expires_at,
			users.id,
			users.email,
			users.password_hash
		FROM password_resets
			JOIN users ON users.id = password_resets.user_id
		WHERE password_resets.token_hash = $1;`,
		tokenHash,
	)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("models.PasswordResetService.Consume")
	}
	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}

	err = service.delete(pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("models.PasswordResetService.Consume: %w", err)
	}

	return &user, nil
}

func (service *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (service *PasswordResetService) delete(id int) error {
	_, err := service.DB.Exec(`
	DELETE FROM password_resets
	WHERE id = $1;
	`, id)
	if err != nil {
		return fmt.Errorf("models.PasswordResetService.delete: %w", err)
	}
	return nil
}
