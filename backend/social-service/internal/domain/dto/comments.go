package dto

type CreateCommentDTO struct {
	Body     string  `json:"body" binding:"required,min=1,max=2000"`
	ParentID *string `json:"parent_id,omitempty" binding:"omitnil"`
}

type UpdateCommentDTO struct {
	Body string `json:"body" binding:"required,min=1,max=2000"`
}

type ReportCommentDTO struct {
	Cause string `json:"cause" binding:"required,min=1,max=255"`
}
