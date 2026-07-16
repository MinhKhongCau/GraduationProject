package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"forum-service/configs"
	apihttp "forum-service/internal/api/http"
	"forum-service/internal/api/http/handlers"
	"forum-service/internal/app/service"
	"forum-service/internal/repository/dao"
	"forum-service/internal/repository/db"
)

func main() {
	// 1. Load configuration.
	configs.LoadConfig()
	cfg := configs.AppConfig

	// 2. Connect to Postgres, then run golang-migrate migrations (schema +
	// seed data) before any query runs — see internal/repository/db/migrate.go
	// and SPEC.md §6.1 for why this replaces GORM AutoMigrate here.
	gormDB := db.ConnectDB(cfg)
	db.RunMigrations(cfg)

	// 3. Connect to RabbitMQ (best-effort — falls back to console-log mode
	// if the broker is unreachable, see internal/repository/db/rabbitmq.go).
	db.InitRabbitMQ(cfg)

	// 4. Wire repositories -> services -> handlers.
	categoryRepo := dao.NewCategoryRepository(gormDB)
	postRepo := dao.NewPostRepository(gormDB)
	tagRepo := dao.NewTagRepository(gormDB)
	commentRepo := dao.NewCommentRepository(gormDB)
	likeRepo := dao.NewPostLikeRepository(gormDB)
	bookmarkRepo := dao.NewPostBookmarkRepository(gormDB)

	events := service.NewEventPublisher()

	categoryService := service.NewCategoryService(categoryRepo)
	postService := service.NewPostService(gormDB, postRepo, categoryRepo, tagRepo, events)
	tagService := service.NewTagService(tagRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, events)
	likeService := service.NewLikeService(likeRepo, postRepo, events)
	bookmarkService := service.NewBookmarkService(bookmarkRepo, postRepo, tagRepo)

	h := apihttp.Handlers{
		Category: handlers.NewCategoryHandler(categoryService),
		Post:     handlers.NewPostHandler(postService),
		Tag:      handlers.NewTagHandler(tagService, postService),
		Comment:  handlers.NewCommentHandler(commentService),
		Like:     handlers.NewLikeHandler(likeService),
		Bookmark: handlers.NewBookmarkHandler(bookmarkService),
	}

	// 5. Router + health check.
	r := gin.Default()
	apihttp.SetupRoutes(r, h)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "forum-service",
			"status":  "up and running",
			"db":      "connected",
		})
	})

	// 6. Start server.
	log.Printf("forum-service: listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("forum-service: failed to start server: %v", err)
	}
}
