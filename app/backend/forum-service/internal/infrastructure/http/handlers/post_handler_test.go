package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"forum-service/internal/application/forum"
	"forum-service/internal/infrastructure/http/handlers"
	"forum-service/internal/infrastructure/http/routes"
	"forum-service/internal/infrastructure/messaging"
	"forum-service/internal/infrastructure/persistence/models"
	"forum-service/internal/infrastructure/persistence/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	err = db.AutoMigrate(
		&models.CategoryDAO{},
		&models.PostDAO{},
		&models.CommentDAO{},
		&models.TagDAO{},
		&models.PostTagDAO{},
		&models.PostLikeDAO{},
		&models.PostBookmarkDAO{},
	)
	if err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	return db
}

func setupTestRouter(db *gorm.DB) (*gin.Engine, *forum.PostService, *forum.CommentService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	events := messaging.NewEventPublisher()

	postRepo := repository.NewPostRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewPostLikeRepository(db)
	bookmarkRepo := repository.NewPostBookmarkRepository(db)

	postSvc := forum.NewPostService(db, postRepo, categoryRepo, tagRepo, events)
	commentSvc := forum.NewCommentService(commentRepo, postRepo, events)
	tagSvc := forum.NewTagService(tagRepo)
	categorySvc := forum.NewCategoryService(categoryRepo)
	likeSvc := forum.NewLikeService(likeRepo, postRepo, events)
	bookmarkSvc := forum.NewBookmarkService(bookmarkRepo, postRepo, tagRepo)

	postHandler := handlers.NewPostHandler(postSvc)
	commentHandler := handlers.NewCommentHandler(commentSvc)
	categoryHandler := handlers.NewCategoryHandler(categorySvc)
	tagHandler := handlers.NewTagHandler(tagSvc, postSvc)
	likeHandler := handlers.NewLikeHandler(likeSvc)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkSvc)

	routes.SetupRoutes(r, routes.Handlers{
		Category: categoryHandler,
		Post:     postHandler,
		Tag:      tagHandler,
		Comment:  commentHandler,
		Like:     likeHandler,
		Bookmark: bookmarkHandler,
	})

	return r, postSvc, commentSvc
}

func TestPostHandler(t *testing.T) {
	db := setupTestDB(t)
	r, _, _ := setupTestRouter(db)

	// Seed Category
	cat := models.CategoryDAO{ID: 1, Name: "Học Đường", Slug: "hoc-duong", Description: "Tâm lý học đường", CreatedAt: time.Now()}
	db.FirstOrCreate(&cat, models.CategoryDAO{ID: 1})

	t.Run("TC-FORUM-QNA-01 - Đăng bài viết mới thành công (Happy Case)", func(t *testing.T) {
		body := map[string]interface{}{
			"categoryId": 1,
			"title":      "Làm sao để vượt qua căng thẳng học đường?",
			"content":    "Nội dung chi tiết câu hỏi thảo luận học đường...",
			"tags":       []string{"tam-ly", "hoc-duong"},
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-FORUM-QNA-02 - Đăng bài viết thất bại do thiếu tiêu đề hoặc nội dung", func(t *testing.T) {
		body := map[string]interface{}{
			"categoryId": 1,
			"title":      "",
			"content":    "",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-QNA-03 - Đăng bài viết thất bại do categoryId không tồn tại", func(t *testing.T) {
		body := map[string]interface{}{
			"categoryId": 99999,
			"title":      "Tiêu đề câu hỏi",
			"content":    "Nội dung bài viết",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-QNA-04 - Đăng bài viết thất bại do thiếu header định danh người dùng", func(t *testing.T) {
		body := map[string]interface{}{
			"categoryId": 1,
			"title":      "Tiêu đề câu hỏi",
			"content":    "Nội dung bài viết",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		// Không set X-User-Id

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-QNA-05 - Truy vấn danh sách bài viết theo danh mục và từ khóa", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/forum/posts?categoryId=1&search=cang-thang", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-QNA-06 - Xem chi tiết bài viết theo slug thành công", func(t *testing.T) {
		p := models.PostDAO{
			ID:         10,
			CategoryID: 1,
			AuthorID:   "user-uuid-1",
			Title:      "Bài viết mẫu",
			Slug:       "bai-viet-mau",
			Content:    "Nội dung chi tiết bài viết mẫu",
			Status:     "PUBLISHED",
			CreatedAt:  time.Now(),
		}
		db.Create(&p)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/forum/posts/bai-viet-mau", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-01 - Chủ bài viết chỉnh sửa nội dung bài viết thành công", func(t *testing.T) {
		newTitle := "Tiêu đề đã được chỉnh sửa"
		body := map[string]interface{}{
			"title": newTitle,
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/forum/posts/10", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1") // Chủ bài viết

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-02 - Từ chối chỉnh sửa bài viết của người khác", func(t *testing.T) {
		newTitle := "Hack tiêu đề bài viết người khác"
		body := map[string]interface{}{
			"title": newTitle,
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/forum/posts/10", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "other-user-uuid") // Không phải chủ bài viết

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-03 - Xóa bài viết cá nhân thành công (Soft Delete)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/forum/posts/10", nil)
		req.Header.Set("X-User-Id", "user-uuid-1") // Chủ bài viết

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-04 - Xóa bài viết thất bại do bài viết không tồn tại", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/forum/posts/99999", nil)
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d: %s", w.Code, w.Body.String())
		}
	})
}
