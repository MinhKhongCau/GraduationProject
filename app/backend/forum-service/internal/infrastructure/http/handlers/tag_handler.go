package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/response"
	"forum-service/internal/infrastructure/http/schemas"
	"forum-service/internal/infrastructure/messaging"
	"forum-service/internal/infrastructure/persistence/repository"
)

type TagHandler struct {
	tagSvc  *forum.TagService
	postSvc *forum.PostService
}

func NewTagHandler(tagSvc *forum.TagService, postSvc *forum.PostService) *TagHandler {
	return &TagHandler{tagSvc: tagSvc, postSvc: postSvc}
}

// @Summary      List all tags
// @Description  Get a list of all forum tags.
// @Tags         tags
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/tags [get]
// func (h *TagHandler) List(c *gin.Context) {
func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tagSvc.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list tags", err.Error())
		return
	}
	resp := make([]schemas.TagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, schemas.NewTagResponse(t))
	}
	response.Success(c, "Tags retrieved", resp)
}

// @Summary      List posts by tag slug
// @Description  Get a paginated list of published posts associated with a specific tag slug.
// @Tags         tags
// @Produce      json
// @Param        slug      path      string  true   "Tag Slug"
// @Param        page      query     int     false  "Page number"  default(1)
// @Param        pageSize  query     int     false  "Page size"    default(20)
// @Success      200       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /api/v1/forum/tags/{slug}/posts [get]
func (h *TagHandler) ListPosts(c *gin.Context) {
	slug := c.Param("slug")
	if _, err := h.tagSvc.GetBySlug(slug); err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Tag not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get tag", err.Error())
		return
	}

	result, err := h.postSvc.List(repository.PostListFilter{
		TagSlug:  slug,
		Page:     parseIntOrDefault(c.Query("page"), 1),
		PageSize: parseIntOrDefault(c.Query("pageSize"), 20),
	})
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
