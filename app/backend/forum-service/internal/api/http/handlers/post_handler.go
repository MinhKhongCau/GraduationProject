package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/api/dto"
	"forum-service/internal/api/http/middleware"
	"forum-service/internal/api/http/response"
	"forum-service/internal/app/entity"
	"forum-service/internal/app/service"
	"forum-service/internal/repository/dao"
)

type PostHandler struct {
	svc *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) List(c *gin.Context) {
	f := dao.PostListFilter{
		TagSlug:  c.Query("tag"),
		Search:   c.Query("search"),
		Status:   c.Query("status"),
		Page:     parseIntOrDefault(c.Query("page"), 1),
		PageSize: parseIntOrDefault(c.Query("pageSize"), 20),
	}
	if v := c.Query("categoryId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.CategoryID = &id
		}
	}
	if v := c.Query("authorId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.AuthorID = &id
		}
	}

	// Only the post's own author (or an ADMIN) may filter by a
	// non-PUBLISHED status (SPEC.md §3.4); anyone else is silently
	// restricted back to PUBLISHED.
	if f.Status != "" && f.Status != string(entity.PostStatusPublished) {
		userID, ok := middleware.UserID(c)
		isOwnerQuery := ok && f.AuthorID != nil && *f.AuthorID == userID
		if !middleware.IsAdmin(c) && !isOwnerQuery {
			f.Status = string(entity.PostStatusPublished)
		}
	}

	result, err := h.svc.List(f)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list posts", err.Error())
		return
	}

	items := make([]dto.PostResponse, 0, len(result.Items))
	for _, p := range result.Items {
		items = append(items, dto.NewPostResponse(p))
	}
	response.Success(c, "Posts retrieved", dto.PostListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
	})
}

// GetBySlug handles GET /posts/:id — the path param is named "id" (not
// "slug") to share Gin's GET route tree with GET /posts/:id/comments;
// see router.go's doc comment for why.
func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("id")
	userID, _ := middleware.UserID(c)
	post, err := h.svc.GetBySlug(slug, userID, middleware.IsAdmin(c))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get post", err.Error())
		return
	}
	response.Success(c, "Post retrieved", dto.NewPostDetailResponse(*post))
}

func (h *PostHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	status := entity.PostStatus(req.Status)
	if status == "" {
		status = entity.PostStatusPublished
	}

	post, err := h.svc.Create(service.CreatePostInput{
		CategoryID: req.CategoryID, AuthorID: userID, Title: req.Title, Summary: req.Summary,
		Content: req.Content, ThumbnailURL: req.ThumbnailURL, Status: status, Tags: req.Tags,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusBadRequest, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}
	response.Created(c, "Post created", dto.NewPostDetailResponse(*post))
}

func (h *PostHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	post, err := h.svc.Update(id, userID, middleware.IsAdmin(c), service.UpdatePostInput{
		CategoryID: req.CategoryID, Title: req.Title, Summary: req.Summary,
		Content: req.Content, ThumbnailURL: req.ThumbnailURL, Tags: req.Tags,
	})
	if err != nil {
		writePostServiceError(c, err)
		return
	}
	response.Success(c, "Post updated", dto.NewPostDetailResponse(*post))
}

func (h *PostHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)
	if err := h.svc.SoftDelete(id, userID, middleware.IsAdmin(c)); err != nil {
		writePostServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PostHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req dto.ChangePostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	post, err := h.svc.ChangeStatus(id, userID, middleware.IsAdmin(c), entity.PostStatus(req.Status))
	if err != nil {
		if errors.Is(err, service.ErrInvalidTransition) {
			response.Error(c, http.StatusBadRequest, "Invalid status transition", err.Error())
			return
		}
		writePostServiceError(c, err)
		return
	}
	response.Success(c, "Post status updated", dto.NewPostDetailResponse(*post))
}

func writePostServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, http.StatusForbidden, "You do not have permission to modify this post", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Post operation failed", err.Error())
	}
}

func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
