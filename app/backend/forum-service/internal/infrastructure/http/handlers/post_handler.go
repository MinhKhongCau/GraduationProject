package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/application/forum"
	postdomain "forum-service/internal/domain/post"
	"forum-service/internal/infrastructure/http/middleware"
	"forum-service/internal/infrastructure/http/response"
	"forum-service/internal/infrastructure/http/schemas"
	"forum-service/internal/infrastructure/messaging"
	"forum-service/internal/infrastructure/persistence/repository"
)

type PostHandler struct {
	svc *forum.PostService
}

func NewPostHandler(svc *forum.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// @Summary      List and search posts
// @Description  Retrieve a paginated list of posts with filtering. Only PUBLISHED posts are returned unless the caller is the author or ADMIN.
// @Tags         posts
// @Produce      json
// @Param        page        query     int     false  "Page number"      default(1)
// @Param        pageSize    query     int     false  "Page size"        default(20)
// @Param        tag         query     string  false  "Filter by tag slug"
// @Param        search      query     string  false  "Search in title/summary"
// @Param        status      query     string  false  "Filter by status (DRAFT/PUBLISHED/ARCHIVED)"
// @Param        categoryId  query     int     false  "Filter by category ID"
// @Param        authorId    query     int     false  "Filter by author ID"
// @Success      200         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /api/v1/forum/posts [get]
func (h *PostHandler) List(c *gin.Context) {
	f := repository.PostListFilter{
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
		f.AuthorID = &v
	}

	// Only the post's own author (or an ADMIN) may filter by a
	// non-PUBLISHED status (SPEC.md §3.4); anyone else is silently
	// restricted back to PUBLISHED.
	if f.Status != "" && f.Status != string(postdomain.PostStatusPublished) {
		userID, ok := middleware.UserID(c)
		isOwnerQuery := ok && f.AuthorID != nil && *f.AuthorID == userID
		if !middleware.IsAdmin(c) && !isOwnerQuery {
			f.Status = string(postdomain.PostStatusPublished)
		}
	}

	result, err := h.svc.List(f)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list posts", err.Error())
		return
	}

	authorIDsMap := make(map[string]bool)
	for _, p := range result.Items {
		if p.AuthorID != "" {
			authorIDsMap[p.AuthorID] = true
		}
	}
	authorIDs := make([]string, 0, len(authorIDsMap))
	for id := range authorIDsMap {
		authorIDs = append(authorIDs, id)
	}

	profiles, err := messaging.FetchProfiles(authorIDs)
	if err != nil {
		log.Printf("forum-service: WARNING failed to fetch author profiles: %v", err)
	}

	items := make([]schemas.PostResponse, 0, len(result.Items))
	for _, p := range result.Items {
		var authorDTO *schemas.AuthorDTO
		if prof, exists := profiles[p.AuthorID]; exists {
			authorDTO = &schemas.AuthorDTO{
				ID:        prof.ID,
				Name:      prof.Name,
				AvatarURL: prof.AvatarURL,
				Role:      prof.Role,
			}
		} else {
			authorDTO = &schemas.AuthorDTO{
				ID:   p.AuthorID,
				Name: "Community member",
				Role: "PATIENT",
			}
		}
		items = append(items, schemas.NewPostResponse(p, authorDTO))
	}
	response.Success(c, "Posts retrieved", schemas.PostListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
	})
}

// @Summary      Get post by slug
// @Description  Retrieve full post detail by its slug and increment view count.
// @Tags         posts
// @Produce      json
// @Param        id   path      string  true  "Post Slug"
// @Success      200  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id} [get]
func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("id")
	userID, _ := middleware.UserID(c)
	post, err := h.svc.GetBySlug(slug, userID, middleware.IsAdmin(c))
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get post", err.Error())
		return
	}
	response.Success(c, "Post retrieved", schemas.NewPostDetailResponse(*post, getAuthorDTO(post.AuthorID)))
}

// @Summary      Create post
// @Description  Create a new post. Tags supplied by name that do not exist are auto-created.
// @Tags         posts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.CreatePostRequest  true  "Post creation payload"
// @Success      201      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req schemas.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	status := postdomain.PostStatus(req.Status)
	if status == "" {
		status = postdomain.PostStatusPublished
	}

	post, err := h.svc.Create(forum.CreatePostInput{
		CategoryID: req.CategoryID, AuthorID: userID, Title: req.Title, Summary: req.Summary,
		Content: req.Content, ThumbnailURL: req.ThumbnailURL, Status: status, Tags: req.Tags,
	})
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusBadRequest, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}
	response.Created(c, "Post created", schemas.NewPostDetailResponse(*post, getAuthorDTO(post.AuthorID)))
}

// @Summary      Update post
// @Description  Update an existing post. Restricted to post Owner or ADMIN.
// @Tags         posts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Post ID"
// @Param        request  body      schemas.UpdatePostRequest  true  "Post update payload"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/posts/{id} [put]
func (h *PostHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req schemas.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	post, err := h.svc.Update(id, userID, middleware.IsAdmin(c), forum.UpdatePostInput{
		CategoryID: req.CategoryID, Title: req.Title, Summary: req.Summary,
		Content: req.Content, ThumbnailURL: req.ThumbnailURL, Tags: req.Tags,
	})
	if err != nil {
		writePostServiceError(c, err)
		return
	}
	response.Success(c, "Post updated", schemas.NewPostDetailResponse(*post, getAuthorDTO(post.AuthorID)))
}

// @Summary      Soft-delete post
// @Description  Soft delete a post. Comments and likes remain in DB but are excluded from read endpoints. Restricted to post Owner or ADMIN.
// @Tags         posts
// @Security     BearerAuth
// @Param        id   path      int  true  "Post ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id} [delete]
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

// @Summary      Change post status
// @Description  Transition a post's status (DRAFT -> PUBLISHED, or PUBLISHED/DRAFT -> ARCHIVED). Restricted to post Owner or ADMIN.
// @Tags         posts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "Post ID"
// @Param        request  body      schemas.ChangePostStatusRequest  true  "Status change payload"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/status [patch]
func (h *PostHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	var req schemas.ChangePostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	post, err := h.svc.ChangeStatus(id, userID, middleware.IsAdmin(c), postdomain.PostStatus(req.Status))
	if err != nil {
		if errors.Is(err, forum.ErrInvalidTransition) {
			response.Error(c, http.StatusBadRequest, "Invalid status transition", err.Error())
			return
		}
		writePostServiceError(c, err)
		return
	}
	response.Success(c, "Post status updated", schemas.NewPostDetailResponse(*post, getAuthorDTO(post.AuthorID)))
}

func writePostServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, forum.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
	case errors.Is(err, forum.ErrForbidden):
		response.Error(c, http.StatusForbidden, "You do not have permission to modify this post", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Post operation failed", err.Error())
	}
}

func getAuthorDTO(authorID string) *schemas.AuthorDTO {
	if authorID == "" {
		return nil
	}
	profiles, err := messaging.FetchProfiles([]string{authorID})
	if err != nil {
		return &schemas.AuthorDTO{
			ID:   authorID,
			Name: "Community member",
			Role: "PATIENT",
		}
	}
	if prof, exists := profiles[authorID]; exists {
		return &schemas.AuthorDTO{
			ID:        prof.ID,
			Name:      prof.Name,
			AvatarURL: prof.AvatarURL,
			Role:      prof.Role,
		}
	}
	return &schemas.AuthorDTO{
		ID:   authorID,
		Name: "Community member",
		Role: "PATIENT",
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
