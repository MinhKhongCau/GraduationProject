// File: internal/infrastructure/http/response/response.go
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// timestampLayout là ISO-8601 UTC có mili-giây, VD: 2026-10-06T09:33:07.151Z.
const timestampLayout = "2006-01-02T15:04:05.000Z07:00"

// BaseResponse là envelope chung cho MỌI response của profile-service (kể cả lỗi):
//
//	{ "message": "...", "statusCode": 200, "timestamp": "...", "result": {...} }
type BaseResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Timestamp  string `json:"timestamp"`
	Result     any    `json:"result"`
}

// PageResult là "result" của các API danh sách có phân trang.
type PageResult[T any] struct {
	Items       []T   `json:"items"`
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	PageSize    int   `json:"pageSize"`
	TotalPages  int   `json:"totalPages"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

// ErrorResult là "result" của response lỗi, mang chi tiết kỹ thuật của lỗi.
type ErrorResult struct {
	Error string `json:"error"`
}

// NewPageResult tính các trường phân trang dẫn xuất (totalPages, hasNext, hasPrevious).
func NewPageResult[T any](items []T, total int64, page, pageSize int) PageResult[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return PageResult[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

func New(httpStatus int, message string, result any) BaseResponse {
	return BaseResponse{
		Message:    message,
		StatusCode: httpStatus,
		Timestamp:  time.Now().UTC().Format(timestampLayout),
		Result:     result,
	}
}

func JSON(c *gin.Context, httpStatus int, message string, result any) {
	c.JSON(httpStatus, New(httpStatus, message, result))
}

func Success(c *gin.Context, message string, result any) {
	JSON(c, http.StatusOK, message, result)
}

func Created(c *gin.Context, message string, result any) {
	JSON(c, http.StatusCreated, message, result)
}

func Error(c *gin.Context, httpStatus int, message string, err string) {
	JSON(c, httpStatus, message, ErrorResult{Error: err})
}

// NoRoute / NoMethod / Recovery giữ cho cả 404, 405 và panic cũng theo BaseResponse.

func NoRoute(c *gin.Context) {
	Error(c, http.StatusNotFound, "Không tìm thấy tài nguyên", "route not found")
}

func NoMethod(c *gin.Context) {
	Error(c, http.StatusMethodNotAllowed, "Phương thức không được hỗ trợ", "method not allowed")
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, _ any) {
		Error(c, http.StatusInternalServerError, "Lỗi hệ thống", "internal server error")
		c.Abort()
	})
}
