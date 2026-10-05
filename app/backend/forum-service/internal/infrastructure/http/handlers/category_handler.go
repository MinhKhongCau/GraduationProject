package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/response"
	"forum-service/internal/infrastructure/http/schemas"
)

type CategoryHandler struct {
	svc *forum.CategoryService
}

func NewCategoryHandler(svc *forum.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// @Summary      List categories
// @Description  Get all forum categories.
// @Tags         categories
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.svc.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list categories", err.Error())
		return
	}
	resp := make([]schemas.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, schemas.NewCategoryResponse(cat))
	}
	response.Success(c, "Categories retrieved", resp)
}

// @Summary      Get category by slug
// @Description  Get a single forum category details by its slug.
// @Tags         categories
// @Produce      json
// @Param        slug  path      string  true  "Category Slug"
// @Success      200   {object}  response.Response
// @Failure      404   {object}  response.Response
// @Failure      500   {object}  response.Response
// @Router       /api/v1/forum/categories/{slug} [get]
func (h *CategoryHandler) GetBySlug(c *gin.Context) {
	cat, err := h.svc.GetBySlug(c.Param("slug"))
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}
	response.Success(c, "Category retrieved", schemas.NewCategoryResponse(*cat))
}

// @Summary      Create category
// @Description  Create a new category. Restricted to ADMIN.
// @Tags         categories
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.CategoryRequest  true  "Category details"
// @Success      201      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req schemas.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	cat, err := h.svc.Create(req.Name, req.Slug, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}
	response.Created(c, "Category created", schemas.NewCategoryResponse(*cat))
}

// @Summary      Update category
// @Description  Update category details. Restricted to ADMIN.
// @Tags         categories
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                  true  "Category ID"
// @Param        request  body      schemas.CategoryRequest  true  "Category details"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/forum/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}
	var req schemas.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	cat, err := h.svc.Update(id, req.Name, req.Slug, req.Description)
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}
	response.Success(c, "Category updated", schemas.NewCategoryResponse(*cat))
}

// @Summary      Delete category
// @Description  Delete a category. Fails if category has posts. Restricted to ADMIN.
// @Tags         categories
// @Security     BearerAuth
// @Param        id   path      int  true  "Category ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}
	if err := h.svc.Delete(id); err != nil {
		switch {
		case errors.Is(err, forum.ErrNotFound):
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
		case errors.Is(err, forum.ErrCategoryHasPosts):
			response.Error(c, http.StatusConflict, "Category still has posts", "CATEGORY_HAS_POSTS")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}
