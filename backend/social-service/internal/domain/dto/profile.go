package dto

type UserProfileDTO struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	UserName  string `json:"username" validate:"omitempty,username"`
	Bio       string `json:"bio" validate:"omitempty,max=4096"`
	AvatarUrl string `json:"avatar_url" validate:"omitempty,url,startswith=https://,max=2048"`
}

type UpdateProfileDTO struct {
	UserName  *string `json:"username" binding:"omitnil,username"`
	Bio       *string `json:"bio" binding:"omitnil,max=4096"`
	AvatarUrl *string `json:"avatar_url" binding:"omitnil,url,startswith=https://,max=2048"`
}
