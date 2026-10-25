package dto

type ReportDTO struct {
	Cause string `json:"cause" binding:"required,min=1,max=1024"`
}
