package services

import (
	"time"
)

const (
	accessTTL       = 5 * time.Minute
	refreshTTL      = (24 * 30 * 12) * time.Hour
	verificationTTL = 7 * time.Minute
	resetPassTTL    = 7 * time.Minute
	codeSubject     = "Verification Code"
	resetSubject    = "Password Reset Code"
)
