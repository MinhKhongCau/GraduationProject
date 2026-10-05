package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"forum-service/config"
	_ "forum-service/docs"
	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/handlers"
	"forum-service/internal/infrastructure/http/routes"
	"forum-service/internal/infrastructure/messaging"
	"forum-service/internal/infrastructure/persistence"
	"forum-service/internal/infrastructure/persistence/repository"
)

// @title Forum Service API
// @version 1.0
// @description Forum Service API - MindCare
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load configuration.
	config.LoadConfig()
	cfg := config.AppConfig

	// 2. Connect to Postgres, then run golang-migrate migrations (schema +
	// seed data) before any query runs — see internal/infrastructure/persistence/migrate.go
	// and SPEC.md §6.1 for why this replaces GORM AutoMigrate here.
	gormDB := config.ConnectDB(cfg)
	persistence.RunMigrations(cfg)

	// 3. Connect to RabbitMQ (best-effort — falls back to console-log mode
	// if the broker is unreachable, see internal/infrastructure/messaging/rabbitmq.go).
	messaging.InitRabbitMQ(cfg)

	// 4. Wire repositories -> services -> handlers.
	categoryRepo := repository.NewCategoryRepository(gormDB)
	postRepo := repository.NewPostRepository(gormDB)
	tagRepo := repository.NewTagRepository(gormDB)
	commentRepo := repository.NewCommentRepository(gormDB)
	likeRepo := repository.NewPostLikeRepository(gormDB)
	bookmarkRepo := repository.NewPostBookmarkRepository(gormDB)

	events := messaging.NewEventPublisher()

	categoryService := forum.NewCategoryService(categoryRepo)
	postService := forum.NewPostService(gormDB, postRepo, categoryRepo, tagRepo, events)
	tagService := forum.NewTagService(tagRepo)
	commentService := forum.NewCommentService(commentRepo, postRepo, events)
	likeService := forum.NewLikeService(likeRepo, postRepo, events)
	bookmarkService := forum.NewBookmarkService(bookmarkRepo, postRepo, tagRepo)

	h := routes.Handlers{
		Category: handlers.NewCategoryHandler(categoryService),
		Post:     handlers.NewPostHandler(postService),
		Tag:      handlers.NewTagHandler(tagService, postService),
		Comment:  handlers.NewCommentHandler(commentService),
		Like:     handlers.NewLikeHandler(likeService),
		Bookmark: handlers.NewBookmarkHandler(bookmarkService),
	}

	// 5. Router + health check.
	r := gin.Default()
	routes.SetupRoutes(r, h)

	// Swagger endpoint
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
