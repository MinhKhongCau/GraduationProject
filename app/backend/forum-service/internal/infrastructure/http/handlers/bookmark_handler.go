package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/middleware"
	"forum-service/internal/infrastructure/http/response"
	"forum-service/internal/infrastructure/http/schemas"
	"forum-service/internal/infrastructure/messaging"
)

type BookmarkHandler struct {
	svc *forum.BookmarkService
}

func NewBookmarkHandler(svc *forum.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{svc: svc}
}

// @Summary      Bookmark a post
// @Description  Bookmark a post for the authenticated user.
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/bookmark [post]
func (h *BookmarkHandler) Bookmark(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.Bookmark(postID, userID)
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to bookmark post", err.Error())
		return
	}
	response.Success(c, "Post bookmarked", schemas.BookmarkResponse{PostID: postID, Bookmarked: true, BookmarkCount: count})
}

// @Summary      Remove bookmark
// @Description  Remove a post bookmark for the authenticated user.
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/posts/{id}/bookmark [delete]
func (h *BookmarkHandler) Remove(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post id", err.Error())
		return
	}
	userID, _ := middleware.UserID(c)

	count, err := h.svc.RemoveBookmark(postID, userID)
	if err != nil {
		if errors.Is(err, forum.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "Post not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to remove bookmark", err.Error())
		return
	}
	response.Success(c, "Bookmark removed", schemas.BookmarkResponse{PostID: postID, Bookmarked: false, BookmarkCount: count})
}

// @Summary      List my bookmarks
// @Description  Get a paginated list of bookmarked posts of the authenticated user.
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        page      query  int  false  "Page number"  default(1)
// @Param        pageSize  query  int  false  "Page size"    default(20)
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/forum/users/me/bookmarks [get]
func (h *BookmarkHandler) ListMine(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	page := parseIntOrDefault(c.Query("page"), 1)
	pageSize := parseIntOrDefault(c.Query("pageSize"), 20)

	result, err := h.svc.ListMine(userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list bookmarks", err.Error())
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
	response.Success(c, "Bookmarks retrieved", schemas.PostListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
	})
}
