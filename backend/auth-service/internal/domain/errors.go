package domain

import "errors"

var ErrEmptyToken = errors.New("missing token")
var ErrNotFound = errors.New("not found")
var ErrInvalidCredentials = errors.New("incorrect email or password")
var ErrEmptyPayload = errors.New("empty payload")
var ErrPasswordMismatch = errors.New("the passwords don't match")
var ErrAlreadyExists = errors.New("data already exists")
var ErrInvalidGoogleToken = errors.New("invalid google token")
var ErrInvalidCode = errors.New("incorrect or expired code")
var ErrEmptyValue = errors.New("value is empty")

// resend
var ErrToManyRequest = errors.New("to many request")

// middleware
var ErrValueNotFound = errors.New("value for key not found")
var ErrIncorrectType = errors.New("incorrect type")

// worker
var ErrLimitExceeded = errors.New("waiting limit exceeded")
