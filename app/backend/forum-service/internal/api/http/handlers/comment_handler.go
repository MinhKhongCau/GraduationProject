package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/api/dto"
	"forum-service/internal/api/http/middleware"
	"forum-service/internal/api/http/response"
	"forum-service/internal/app/entity"
	"forum-service/internal/app/service"
	"forum-service/internal/repository/db"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// Tree handles GET /posts/:id/comments — :id here is the numeric post id
// (see router.go's doc comment on why the GET tree shares one param name).
// @Summary      Get comment tree
// @Description  Get a nested comment tree for a specific post.
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/comments [get]
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

	userIDsMap := make(map[string]bool)
	collectUserIDs(roots, userIDsMap)
	userIDs := make([]string, 0, len(userIDsMap))
	for id := range userIDsMap {
		userIDs = append(userIDs, id)
	}

	profiles := getUserProfileMap(userIDs)

	resp := make([]dto.CommentResponse, 0, len(roots))
	for _, r := range roots {
		resp = append(resp, dto.NewCommentResponse(r, profiles))
	}
	response.Success(c, "Comments retrieved", resp)
}

// @Summary      Create comment
// @Description  Create a comment or reply to a comment on a specific post.
// @Tags         comments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "Post ID"
// @Param        request  body      dto.CreateCommentRequest  true  "Comment payload"
// @Success      201      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/comments [post]
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
	response.Created(c, "Comment created", dto.NewCommentResponse(comment, getUserProfileMap([]string{comment.UserID})))
}

// @Summary      Edit comment
// @Description  Edit the content of a comment. Restricted to comment Owner or ADMIN.
// @Tags         comments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "Comment ID"
// @Param        request  body      dto.UpdateCommentRequest  true  "Comment edit payload"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/comments/{id} [put]
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
	response.Success(c, "Comment updated", dto.NewCommentResponse(comment, getUserProfileMap([]string{comment.UserID})))
}

func collectUserIDs(comments []*entity.Comment, ids map[string]bool) {
	for _, c := range comments {
		if c.UserID != "" {
			ids[c.UserID] = true
		}
		collectUserIDs(c.Replies, ids)
	}
}

func getUserProfileMap(userIDs []string) map[string]dto.AuthorDTO {
	result := make(map[string]dto.AuthorDTO)
	if len(userIDs) == 0 {
		return result
	}
	profiles, err := db.FetchProfiles(userIDs)
	if err != nil {
		log.Printf("forum-service: WARNING failed to fetch user profiles: %v", err)
		return result
	}
	for k, v := range profiles {
		result[k] = dto.AuthorDTO{
			ID:        v.ID,
			Name:      v.Name,
			AvatarURL: v.AvatarURL,
			Role:      v.Role,
		}
	}
	return result
}

// @Summary      Soft-delete comment
// @Description  Soft-delete a comment. Replies remain intact but the deleted comment's content is set to null. Restricted to comment Owner or ADMIN.
// @Tags         comments
// @Security     BearerAuth
// @Param        id   path      int  true  "Comment ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/comments/{id} [delete]
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
