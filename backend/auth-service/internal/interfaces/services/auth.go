package services

import (
	"time"
)

const (
	AccessTTL       = 5 * time.Minute
	RefreshTTL      = (24 * 30 * 12) * time.Hour
	VerificationTTL = 7 * time.Minute
	ResetPassTTL    = 7 * time.Minute
	codeSubject     = "Verification Code"
	resetSubject    = "Password Reset Code"
)
