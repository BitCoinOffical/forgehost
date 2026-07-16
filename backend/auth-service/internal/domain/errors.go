package domain

import "errors"

var ErrEmptyToken = errors.New("missing token")
var ErrNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("incorrect email or password")
var ErrEmptyPayload = errors.New("empty payload")
var ErrPasswordMismatch = errors.New("the passwords don't match")
var ErrEmailAlreadyExists = errors.New("such user is already registered")
