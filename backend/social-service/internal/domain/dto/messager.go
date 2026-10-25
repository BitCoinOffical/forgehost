package dto

type MessageDTO struct {
	Text string `json:"text" binding:"required,min=1,max=4096"`
}
