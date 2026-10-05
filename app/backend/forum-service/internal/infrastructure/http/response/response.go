package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard envelope used by every endpoint, matching the
// {success, message, data, error} shape used across this platform's other
// Go services (see e.g. payment-service/pkg/response).
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(c *gin.Context, httpStatus int, success bool, message string, data interface{}, errMsg string) {
	c.JSON(httpStatus, Response{Success: success, Message: message, Data: data, Error: errMsg})
}

func Success(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, true, message, data, "")
}

func Created(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusCreated, true, message, data, "")
}

func Error(c *gin.Context, httpStatus int, message string, err string) {
	JSON(c, httpStatus, false, message, nil, err)
}
