package dto

type ExchangeRequestDTO struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

type TokensDTO struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateTokensDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleAndroidUserDTO struct {
	IDToken string `json:"id_token" binding:"required"`
}

type GoogleUserDTO struct {
	Sub           string `json:"sub"`
	Name          string `json:"name" binding:"required,max=255"`
	GivenName     string `json:"given_name" binding:"max=255"`
	FamilyName    string `json:"family_name" binding:"max=255"`
	Picture       string `json:"picture"`
	Email         string `json:"email" binding:"required,email,max=255"`
	EmailVerified bool   `json:"email_verified"`
}

type UsersRegisterDTO struct {
	Email         string `json:"email" binding:"required,email,max=255"`
	Password      string `json:"password" binding:"required,min=8"`
	PasswordRetry string `json:"password_retry" binding:"required,min=8,eqfield=Password"`
}

type UsersLoginDTO struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8"`
}

type VerifyEmailDTO struct {
	Email      string `json:"email" binding:"required,email,max=255"`
	PendingKey string `json:"pending_key" binding:"required"`
	Code       string `json:"code" binding:"required,len=6,numeric"`
}

type UserPasswordDTO struct {
	OldPassword      string `json:"old_password" binding:"required,min=8"`
	NewPassword      string `json:"new_password" binding:"required,min=8"`
	NewPasswordRetry string `json:"new_password_retry" binding:"required,min=8,max=72,eqfield=NewPassword"`
}

type PasswordResetDTO struct {
	Email            string `json:"email" binding:"required,email,max=255"`
	NewPassword      string `json:"new_password" binding:"required,min=8"`
	NewPasswordRetry string `json:"new_password_retry" binding:"required,min=8,eqfield=NewPassword"`
	PendingKey       string `json:"pending_key" binding:"required"`
	Code             string `json:"code" binding:"required,len=6,numeric"`
}

type PendingKeyDTO struct {
	PendingKey string `json:"pending_key" binding:"required"`
}
