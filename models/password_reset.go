package models

import (
	"database/sql"
	"fmt"
	"time"
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

func (ps *PasswordResetService) Create(email string) (*PasswordReset, error) {
	return nil, fmt.Errorf("")
}

func (ps *PasswordResetService) Consume(token string) (*User, error) {
	return nil, fmt.Errorf("")
}
