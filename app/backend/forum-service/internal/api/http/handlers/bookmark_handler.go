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

type BookmarkHandler struct {
	svc *service.BookmarkService
}

func NewBookmarkHandler(svc *service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{svc: svc}
}

func (h *BookmarkHandler) Bookmark(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Bookmark(postID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to bookmark post", err.Error())
		return
	}
	response.Success(c, "Post bookmarked", dto.BookmarkResponse{PostID: postID, Bookmarked: true, BookmarkCount: count})
}

func (h *BookmarkHandler) Remove(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.RemoveBookmark(postID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to remove bookmark", err.Error())
		return
	}
	response.Success(c, "Bookmark removed", dto.BookmarkResponse{PostID: postID, Bookmarked: false, BookmarkCount: count})
}

func (h *BookmarkHandler) ListMine(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	page := parseIntOrDefault(c.Query("page"), 1)
	pageSize := parseIntOrDefault(c.Query("pageSize"), 20)

	result, err := h.svc.ListMine(userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list bookmarks", err.Error())
		return
	}

	items := make([]dto.PostResponse, 0, len(result.Items))
	for _, p := range result.Items {
		items = append(items, dto.NewPostResponse(p))
	}
	response.Success(c, "Bookmarks retrieved", dto.PostListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
	})
}
