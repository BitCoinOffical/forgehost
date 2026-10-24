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
var ErrForbidden = errors.New("insufficient privileges")
var ErrCannotKickOwner = errors.New("cannot kick owner")
var ErrUserOffline = errors.New("user offline")
var ErrSubscribeToSelf = errors.New("cannot subscribe to yourself")
var ErrReportToSelf = errors.New("cannot report to yourself")
var ErrBlockToSelf = errors.New("cannot block to yourself")
var ErrInvalidCursor = errors.New("invalid cursor")

// resend
var ErrToManyRequest = errors.New("to many request")

// middleware
var ErrValueNotFound = errors.New("value for key not found")
var ErrIncorrectType = errors.New("incorrect type")

// worker
var ErrLimitExceeded = errors.New("waiting limit exceeded")

// stream
var ErrStreamExists = errors.New("stream already exists")
var ErrStreamNotFound = errors.New("stream not found")
var ErrChannelNotLinked = errors.New("the channel is not linked.")
