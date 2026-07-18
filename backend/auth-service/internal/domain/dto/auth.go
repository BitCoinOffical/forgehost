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
