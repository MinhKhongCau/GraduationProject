package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"forum-service/internal/api/dto"
	"forum-service/internal/api/http/response"
	"forum-service/internal/app/service"
	"forum-service/internal/repository/dao"
)

type TagHandler struct {
	tagSvc  *service.TagService
	postSvc *service.PostService
}

func NewTagHandler(tagSvc *service.TagService, postSvc *service.PostService) *TagHandler {
	return &TagHandler{tagSvc: tagSvc, postSvc: postSvc}
}

// @Summary      List all tags
// @Description  Get a list of all forum tags.
// @Tags         tags
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tagSvc.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list tags", err.Error())
		return
	}
	resp := make([]dto.TagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, dto.NewTagResponse(t))
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
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Tag not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get tag", err.Error())
		return
	}

	result, err := h.postSvc.List(dao.PostListFilter{
		TagSlug:  slug,
		Page:     parseIntOrDefault(c.Query("page"), 1),
		PageSize: parseIntOrDefault(c.Query("pageSize"), 20),
	})
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
