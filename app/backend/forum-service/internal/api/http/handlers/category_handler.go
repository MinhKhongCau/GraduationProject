package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/api/dto"
	"forum-service/internal/api/http/response"
	"forum-service/internal/app/service"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.svc.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list categories", err.Error())
		return
	}
	resp := make([]dto.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, dto.NewCategoryResponse(cat))
	}
	response.Success(c, "Categories retrieved", resp)
}

func (h *CategoryHandler) GetBySlug(c *gin.Context) {
	cat, err := h.svc.GetBySlug(c.Param("slug"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}
	response.Success(c, "Category retrieved", dto.NewCategoryResponse(*cat))
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	cat, err := h.svc.Create(req.Name, req.Slug, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}
	response.Created(c, "Category created", dto.NewCategoryResponse(*cat))
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	cat, err := h.svc.Update(id, req.Name, req.Slug, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}
	response.Success(c, "Category updated", dto.NewCategoryResponse(*cat))
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}
	if err := h.svc.Delete(id); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
		case errors.Is(err, service.ErrCategoryHasPosts):
			response.Error(c, http.StatusConflict, "Category still has posts", "CATEGORY_HAS_POSTS")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}
