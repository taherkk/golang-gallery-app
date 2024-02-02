package models

import "errors"

var (
	ErrNotFound   = errors.New("resource could not be found")
	ErrEmailToken = errors.New("email address is already in use")
)
