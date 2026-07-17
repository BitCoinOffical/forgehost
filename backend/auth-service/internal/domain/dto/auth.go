package dto

import (
	"github.com/google/uuid"
)

type TokensDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GoogleUserDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Picture  string `json:"picture"`
}

type UsersRegisterDTO struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	PasswordRetry string `json:"password_retry"`
}

type UsersLoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OAuthAccountDTO struct {
	UserID         uuid.UUID `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	GivenName      string    `json:"given_name"`
	FamilyName     string    `json:"family_name"`
	EmailVerified  bool      `json:"email_verified"`
}
