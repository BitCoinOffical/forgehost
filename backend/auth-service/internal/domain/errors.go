package domain

import "errors"

var ErrEmptyToken = errors.New("missing token")
var ErrNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("incorrect email or password")
