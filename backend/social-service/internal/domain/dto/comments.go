package dto

type CreateCommentDTO struct {
	Body     string  `json:"body"`
	ParentID *string `json:"parent_id,omitempty"`
}

type UpdateCommentDTO struct {
	Body string `json:"body"`
}

type ReportCommentDTO struct {
	Cause string `json:"cause"`
}
