package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(c *gin.Context, httpStatus int, success bool, message string, data interface{}, err string) {
	c.JSON(httpStatus, Response{
		Success: success,
		Message: message,
		Data:    data,
		Error:   err,
	})
}

func Success(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, true, message, data, "")
}

func Error(c *gin.Context, httpStatus int, message string, err string) {
	JSON(c, httpStatus, false, message, nil, err)
}
