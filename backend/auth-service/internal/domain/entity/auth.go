package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Name          string
	Email         string
	Password      string
	PasswordRetry string
	Picture       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OAuthAccount struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	GivenName      string
	FamilyName     string
	EmailVerified  bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
