package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/middleware"
	"forum-service/internal/infrastructure/http/response"
	"forum-service/internal/infrastructure/http/schemas"
)

type LikeHandler struct {
	svc *forum.LikeService
}

func NewLikeHandler(svc *forum.LikeService) *LikeHandler {
	return &LikeHandler{svc: svc}
}

// @Summary      Like a post
// @Description  Like a post for the authenticated user.
// @Tags         likes
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/like [post]
func (h *LikeHandler) Like(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Like(postID, userID)
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to like post", err.Error())
		return
	}
	response.Success(c, "Post liked", schemas.LikeResponse{PostID: postID, Liked: true, LikeCount: count})
}

// @Summary      Unlike a post
// @Description  Remove a post like for the authenticated user.
// @Tags         likes
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/like [delete]
func (h *LikeHandler) Unlike(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Unlike(postID, userID)
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to unlike post", err.Error())
		return
	}
	response.Success(c, "Post unliked", schemas.LikeResponse{PostID: postID, Liked: false, LikeCount: count})
}
