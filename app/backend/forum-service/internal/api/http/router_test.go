package http_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	apihttp "forum-service/internal/api/http"
	"forum-service/internal/api/http/handlers"
)

// TestSetupRoutes_DoesNotPanic guards against the Gin route-tree wildcard
// collision documented in router.go's doc comment: Gin panics at
// registration time if two routes sharing a tree position (e.g.
// GET /posts/:id and GET /posts/:id/comments) declare different param
// names. Handler fields are left as typed nil pointers — SetupRoutes only
// takes method values, it never calls them, so this exercises just the
// route-tree construction.
func TestSetupRoutes_DoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("SetupRoutes panicked: %v", rec)
		}
	}()

	apihttp.SetupRoutes(r, apihttp.Handlers{
		Category: (*handlers.CategoryHandler)(nil),
		Post:     (*handlers.PostHandler)(nil),
		Tag:      (*handlers.TagHandler)(nil),
		Comment:  (*handlers.CommentHandler)(nil),
		Like:     (*handlers.LikeHandler)(nil),
		Bookmark: (*handlers.BookmarkHandler)(nil),
	})
}
