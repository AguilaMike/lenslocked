package models

import "github.com/AguilaMike/lenslocked/pkg/app/errors"

var (
	ErrNotFound      = errors.New("models: resource could not be found")
	ErrEmailTaken    = errors.Public(errors.New("models: email address is already in use"), "That email address is already associated with an account.")
	ErrUserNotFound  = errors.Public(errors.New("models: we were unable to find a user"), "We were unable to find a user with that email address.")
	ErrPasswordError = errors.Public(errors.New("models: that password is incorrect"), "That password is incorrect.")
	ErrInvalidID     = errors.Public(errors.New("models: ID provided was invalid"), "Invalid ID provided.")
)
