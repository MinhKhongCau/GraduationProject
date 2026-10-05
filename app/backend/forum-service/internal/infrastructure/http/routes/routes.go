package routes

import (
	"github.com/gin-gonic/gin"

	"forum-service/internal/infrastructure/http/handlers"
	"forum-service/internal/infrastructure/http/middleware"
)

type Handlers struct {
	Category *handlers.CategoryHandler
	Post     *handlers.PostHandler
	Tag      *handlers.TagHandler
	Comment  *handlers.CommentHandler
	Like     *handlers.LikeHandler
	Bookmark *handlers.BookmarkHandler
}

// SetupRoutes wires every endpoint from README.md's API Summary Table.
//
// Gin builds one radix tree per HTTP method and panics at startup if two
// routes sharing a tree position use different wildcard parameter names.
// GET /posts/:id (post detail, semantically a slug) and
// GET /posts/:id/comments (semantically a numeric post id) sit at the same
// position in the GET tree, so both — and every other GET route under
// /posts/ — use the single param name "id"; each handler parses it per its
// own contract instead of using distinct names like :slug/:postId.
func SetupRoutes(r *gin.Engine, h Handlers) {
	api := r.Group("/api/v1/forum")
	{
		categories := api.Group("/categories")
		{
			categories.GET("", h.Category.List)
			categories.GET("/:slug", h.Category.GetBySlug)
			categories.POST("", middleware.RequireAuth(), middleware.RequireRole("ADMIN"), h.Category.Create)
			categories.PUT("/:id", middleware.RequireAuth(), middleware.RequireRole("ADMIN"), h.Category.Update)
			categories.DELETE("/:id", middleware.RequireAuth(), middleware.RequireRole("ADMIN"), h.Category.Delete)
		}

		posts := api.Group("/posts")
		{
			posts.GET("", h.Post.List)
			posts.GET("/:id", h.Post.GetBySlug)
			posts.POST("", middleware.RequireAuth(), h.Post.Create)
			posts.PUT("/:id", middleware.RequireAuth(), h.Post.Update)
			posts.DELETE("/:id", middleware.RequireAuth(), h.Post.SoftDelete)
			posts.PATCH("/:id/status", middleware.RequireAuth(), h.Post.ChangeStatus)

			posts.GET("/:id/comments", h.Comment.Tree)
			posts.POST("/:id/comments", middleware.RequireAuth(), h.Comment.Create)

			posts.POST("/:id/like", middleware.RequireAuth(), h.Like.Like)
			posts.DELETE("/:id/like", middleware.RequireAuth(), h.Like.Unlike)

			posts.POST("/:id/bookmark", middleware.RequireAuth(), h.Bookmark.Bookmark)
			posts.DELETE("/:id/bookmark", middleware.RequireAuth(), h.Bookmark.Remove)
		}

		tags := api.Group("/tags")
		{
			tags.GET("", h.Tag.List)
			tags.GET("/:slug/posts", h.Tag.ListPosts)
		}

		comments := api.Group("/comments")
		{
			comments.PUT("/:id", middleware.RequireAuth(), h.Comment.Edit)
			comments.DELETE("/:id", middleware.RequireAuth(), h.Comment.SoftDelete)
		}

		users := api.Group("/users/me")
		{
			users.GET("/bookmarks", middleware.RequireAuth(), h.Bookmark.ListMine)
		}
	}
}
