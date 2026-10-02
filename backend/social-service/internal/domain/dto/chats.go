package dto

type MemberRole string

type CreateChatDTO struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type UpdateChatDTO struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type SetMemberRoleDTO struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
