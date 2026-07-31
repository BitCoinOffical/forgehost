package models

import (
	"time"

	"github.com/google/uuid"
)

type VerifyEmail struct {
	Email string
	Code  string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type User struct {
	ID            uuid.UUID `db:"id"`
	Name          *string   `db:"name"`
	Email         string    `db:"email"`
	PasswordHash  []byte    `db:"password_hash"`
	Picture       *string   `db:"picture"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type OAuthAccount struct {
	ID             uuid.UUID `db:"id"`
	UserID         uuid.UUID `db:"user_id"`
	Provider       string    `db:"provider"`
	ProviderUserID string    `db:"provider_user_id"` //sub
	GivenName      string    `db:"given_name"`
	FamilyName     string    `db:"family_name"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
