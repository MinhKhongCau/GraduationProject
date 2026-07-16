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

type LikeHandler struct {
	svc *service.LikeService
}

func NewLikeHandler(svc *service.LikeService) *LikeHandler {
	return &LikeHandler{svc: svc}
}

func (h *LikeHandler) Like(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Like(postID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to like post", err.Error())
		return
	}
	response.Success(c, "Post liked", dto.LikeResponse{PostID: postID, Liked: true, LikeCount: count})
}

func (h *LikeHandler) Unlike(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Unlike(postID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to unlike post", err.Error())
		return
	}
	response.Success(c, "Post unliked", dto.LikeResponse{PostID: postID, Liked: false, LikeCount: count})
}
