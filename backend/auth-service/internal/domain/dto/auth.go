package dto

type TokensDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GoogleAndroidUserDTO struct {
	IdToken string `json:"id_token"`
}

type GoogleUserDTO struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
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

type VerifyEmailDTO struct {
	Email string `json:"email"`
	PendingKeyDTO
	Code string `json:"code"`
}

type UserPasswordDTO struct {
	OldPassword string `json:"old_password"`
	PasswordPair
}

type PasswordResetDTO struct {
	Email string `json:"email"`
	PasswordPair
	PendingKeyDTO
	Code string `json:"code"`
}

type PasswordPair struct {
	NewPassword      string `json:"new_password"`
	NewPasswordRetry string `json:"new_password_retry"`
}

type PendingKeyDTO struct {
	PendingKey string `json:"pending_key"`
}
