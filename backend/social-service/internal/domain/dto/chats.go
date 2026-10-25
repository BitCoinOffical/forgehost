package dto

type MemberRole string

const (
	RoleMember    MemberRole = "member"
	RoleModerator MemberRole = "moderator"
)

type CreateChatDTO struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description" binding:"omitnil,max=4096"`
}

type UpdateChatDTO struct {
	Title       *string `json:"title" binding:"omitnil,min=1,max=255"`
	Description *string `json:"description" binding:"omitnil,max=4096"`
}

type SetMemberRoleDTO struct {
	UserID string     `json:"user_id" binding:"required"`
	Role   MemberRole `json:"role" binding:"required,oneof=member moderator"`
}
