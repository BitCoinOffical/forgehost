package dto

type MemberRole string

const (
	RoleMember MemberRole = "member"
	RoleAdmin  MemberRole = "moderator"
)

type CreateChatDTO struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type UpdateChatDTO struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type SetMemberRoleDTO struct {
	UserID string     `json:"user_id"`
	Role   MemberRole `json:"role"`
}
