package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/api/dto"
	"forum-service/internal/api/http/middleware"
	"forum-service/internal/api/http/response"
	"forum-service/internal/app/service"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// Tree handles GET /posts/:id/comments — :id here is the numeric post id
// (see router.go's doc comment on why the GET tree shares one param name).
func (h *CommentHandler) Tree(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	roots, err := h.svc.Tree(postID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get comments", err.Error())
		return
	}
	resp := make([]dto.CommentResponse, 0, len(roots))
	for _, r := range roots {
		resp = append(resp, dto.NewCommentResponse(r))
	}
	response.Success(c, "Comments retrieved", resp)
}

func (h *CommentHandler) Create(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	comment, err := h.svc.Create(postID, userID, req.ParentID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post or parent comment not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create comment", err.Error())
		return
	}
	response.Created(c, "Comment created", dto.NewCommentResponse(comment))
}

func (h *CommentHandler) Edit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	comment, err := h.svc.Edit(id, userID, middleware.IsAdmin(c), req.Content)
	if err != nil {
		writeCommentServiceError(c, err)
		return
	}
	response.Success(c, "Comment updated", dto.NewCommentResponse(comment))
}

func (h *CommentHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)
	if err := h.svc.SoftDelete(id, userID, middleware.IsAdmin(c)); err != nil {
		writeCommentServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeCommentServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Comment not found", err.Error())
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, http.StatusForbidden, "You do not have permission to modify this comment", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Comment operation failed", err.Error())
	}
}
