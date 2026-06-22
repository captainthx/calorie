package user

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewUsersHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetUserByToken(c *gin.Context) {
	fmt.Println("GetUserByToken called")
	token := c.Param("token")
	user, err := h.service.GetUserByToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}
