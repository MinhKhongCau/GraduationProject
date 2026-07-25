package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"forum-service/internal/app/entity"
	"forum-service/internal/repository/dao"
)

func TestCommentHandler(t *testing.T) {
	db := setupTestDB(t)
	r, _, _ := setupTestRouter(db)

	// Seed Category & Post
	cat := dao.CategoryDAO{ID: 1, Name: "Thảo luận chung", Slug: "thao-luan", Description: "Mô tả", CreatedAt: time.Now()}
	db.FirstOrCreate(&cat, dao.CategoryDAO{ID: 1})

	post := dao.PostDAO{
		ID:         20,
		CategoryID: 1,
		AuthorID:   "user-author-uuid",
		Title:      "Bài viết thảo luận bình luận",
		Slug:       "bai-viet-thao-luan",
		Content:    "Nội dung bài viết",
		Status:     "PUBLISHED",
		CreatedAt:  time.Now(),
	}
	db.Create(&post)

	t.Run("TC-FORUM-MNG-05 - Người dùng gửi bình luận vào bài viết thành công", func(t *testing.T) {
		body := map[string]interface{}{
			"content": "Cảm ơn tác giả về bài viết rất hữu ích!",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts/20/comments", bytes.NewBuffer(jsonBytes))
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

	t.Run("TC-FORUM-MNG-06 - Gửi bình luận thất bại do nội dung bình luận rỗng", func(t *testing.T) {
		body := map[string]interface{}{
			"content": "",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/posts/20/comments", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-07 - Chỉnh sửa nội dung bình luận cá nhân thành công", func(t *testing.T) {
		// Seed a comment owned by user-uuid-1
		comment := dao.CommentDAO{
			ID:        50,
			PostID:    20,
			UserID:    "user-uuid-1",
			Path:      entity.LTree("50"),
			Content:   "Bình luận ban đầu",
			CreatedAt: time.Now(),
		}
		db.Create(&comment)

		body := map[string]interface{}{
			"content": "Nội dung bình luận đã được chỉnh sửa.",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/forum/comments/50", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("TC-FORUM-MNG-08 - Xóa bình luận cá nhân thành công (Soft Delete)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/forum/comments/50", nil)
		req.Header.Set("X-User-Id", "user-uuid-1")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d: %s", w.Code, w.Body.String())
		}
	})
}
