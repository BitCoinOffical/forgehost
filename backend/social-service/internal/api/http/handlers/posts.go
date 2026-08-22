package handlers

import "github.com/gin-gonic/gin"

type PostHandler struct {
}

func NewPostHandler() *PostHandler {
	return &PostHandler{}
}

func (h *PostHandler) List(c *gin.Context) {

}

func (h *PostHandler) GetByID(c *gin.Context) {

}

func (h *PostHandler) Create(c *gin.Context) {

}

func (h *PostHandler) Update(c *gin.Context) {

}

func (h *PostHandler) Delete(c *gin.Context) {

}

func (h *PostHandler) Report(c *gin.Context) {

}

func (h *PostHandler) Like(c *gin.Context) {

}

func (h *PostHandler) Unlike(c *gin.Context) {

}
