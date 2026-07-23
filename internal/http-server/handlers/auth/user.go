package auth

import "errors"

type User struct {
	ID           int64
	LoginName    string
	PasswordHash string
	Enabled      bool
}

var ErrUserNotFound = errors.New("user not found")
