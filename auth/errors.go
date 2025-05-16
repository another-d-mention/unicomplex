package auth

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupAlreadyExists = errors.New("group already exists")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
