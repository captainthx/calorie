package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const InternalServerErrorMessage = "internal server error"

type SuccessBody struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
}

type ErrorBody struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"forbidden"`
}

type MessageData struct {
	Message string `json:"message" example:"Entry deleted successfully"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessBody{Success: true, Data: data})
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorBody{Success: false, Error: msg})
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, msg)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, msg)
}

func InternalServerError(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}
	Error(c, http.StatusInternalServerError, InternalServerErrorMessage)
}
