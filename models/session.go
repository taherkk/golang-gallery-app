package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/taherk/galleryapp/rand"
)

const MinSessionTokenBytes = 32

type Session struct {
	ID     int
	UserID uint
	// token is only set when creating a new session. When look up a sesstion
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token.
	Token     string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	// how many bytes to use when generating each password reset token. If this vale is
	// not set or is less than the MinBytesPerToken const it will be ignored and
	// MinBytesPerToken will be used.
	BytesPerToken int
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (ss *SessionService) Create(userID uint) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinSessionTokenBytes {
		bytesPerToken = MinSessionTokenBytes
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("session.create: %w", err)
	}

	tokenHash := ss.hash(token)

	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: tokenHash,
	}

	row := ss.DB.QueryRow(`
		INSERT INTO sessions (user_id, token_hash)
		VALUES ($1, $2) ON CONFLICT (user_id) DO
		UPDATE
		SET token_hash = $2
		RETURNING id;`,
		userID, tokenHash,
	)

	err = row.Scan(&session.ID)

	if err != nil {
		return nil, fmt.Errorf("models.session.create: %w", err)
	}

	return &session, nil
}

func (ss *SessionService) User(token string) (*User, error) {

	// hash the session token
	tokenHash := ss.hash(token)

	// query that session with the hash
	row := ss.DB.QueryRow(`
		SELECT users.id, users.email, users.password_hash
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash=$1;`,
		tokenHash,
	)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("models.session.user: %w", err)
	}

	return &user, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)

	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash=$1;
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("models.session.delete: %w", err)
	}

	return nil
}
