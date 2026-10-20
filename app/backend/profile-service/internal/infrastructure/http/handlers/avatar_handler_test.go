package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/infrastructure/http/middleware"
	s3infra "profile-service/internal/infrastructure/s3"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pngBytes là header PNG tối thiểu để http.DetectContentType nhận diện đúng (chỉ dùng trong fake uploader).
var pngBytes = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0x0d}

// fakeAvatarUploader ghi lại lần gọi và trả về object key/lỗi đã cấu hình, không kết nối S3 thật.
type fakeAvatarUploader struct {
	url   string
	err   error
	calls int
}

func (f *fakeAvatarUploader) UploadProfileImage(_ context.Context, _ string, _ io.ReadSeeker) (string, error) {
	f.calls++
	return f.url, f.err
}

// multipartAvatarBody tạo body multipart với một file tại fieldName.
func multipartAvatarBody(t *testing.T, fieldName string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, "avatar.png")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return body, writer.FormDataContentType()
}

// newAvatarContext tạo gin context như RequireAuth đã chạy: caller là authID với role đã cho.
func newAvatarContext(t *testing.T, targetID uuid.UUID, callerID uuid.UUID, role string, body *bytes.Buffer, contentType string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/profiles/"+targetID.String()+"/avatar", body)
	c.Request.Header.Set("Content-Type", contentType)
	c.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	c.Set(middleware.CtxAuthID, callerID.String())
	c.Set(middleware.CtxRole, role)
	return c, w
}

// expectProfileLookup mock truy vấn profile theo auth_id trả về một profile với vai trò đã cho.
func expectProfileLookup(mock sqlmock.Sqlmock, profileID, authID uuid.UUID, role string) {
	mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role"}).AddRow(profileID, authID, role))
}

func TestUploadAvatar_Forbidden_WhenNotOwnerOrAdmin(t *testing.T) {
	SetupTestDB(t)
	uploader := &fakeAvatarUploader{url: "profiles/p/x.png"}
	h := NewAvatarHandler(uploader)

	targetID, callerID := uuid.New(), uuid.New()
	body, ct := multipartAvatarBody(t, "file", pngBytes)
	c, w := newAvatarContext(t, targetID, callerID, "PATIENT", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, uploader.calls)
}

func TestUploadAvatar_ServiceUnavailable_WhenS3NotConfigured(t *testing.T) {
	SetupTestDB(t)
	h := NewAvatarHandler(nil)

	authID := uuid.New()
	body, ct := multipartAvatarBody(t, "file", pngBytes)
	c, w := newAvatarContext(t, authID, authID, "PATIENT", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestUploadAvatar_BadRequest_WhenTargetIsAdmin(t *testing.T) {
	_, mock := SetupTestDB(t)
	uploader := &fakeAvatarUploader{url: "profiles/p/x.png"}
	h := NewAvatarHandler(uploader)

	authID := uuid.New()
	expectProfileLookup(mock, uuid.New(), authID, "ADMIN")
	body, ct := multipartAvatarBody(t, "file", pngBytes)
	c, w := newAvatarContext(t, authID, uuid.New(), "ADMIN", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, uploader.calls)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAvatar_BadRequest_WhenFieldMissing(t *testing.T) {
	_, mock := SetupTestDB(t)
	h := NewAvatarHandler(&fakeAvatarUploader{})

	authID := uuid.New()
	expectProfileLookup(mock, uuid.New(), authID, "PATIENT")
	body, ct := multipartAvatarBody(t, "wrong_field", pngBytes)
	c, w := newAvatarContext(t, authID, authID, "PATIENT", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAvatar_BadRequest_WhenNotImage(t *testing.T) {
	_, mock := SetupTestDB(t)
	h := NewAvatarHandler(&fakeAvatarUploader{err: s3infra.ErrUnsupportedImageType})

	authID := uuid.New()
	expectProfileLookup(mock, uuid.New(), authID, "PATIENT")
	body, ct := multipartAvatarBody(t, "file", []byte("%PDF-1.4 not an image"))
	c, w := newAvatarContext(t, authID, authID, "PATIENT", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAvatar_PayloadTooLarge(t *testing.T) {
	_, mock := SetupTestDB(t)
	uploader := &fakeAvatarUploader{url: "profiles/p/x.png"}
	h := NewAvatarHandler(uploader)

	authID := uuid.New()
	expectProfileLookup(mock, uuid.New(), authID, "PATIENT")
	big := append(append([]byte{}, pngBytes...), bytes.Repeat([]byte{0}, maxAvatarSize)...)
	body, ct := multipartAvatarBody(t, "file", big)
	c, w := newAvatarContext(t, authID, authID, "PATIENT", body, ct)

	h.UploadAvatar(c)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Equal(t, 0, uploader.calls)
}

func TestUploadAvatar_Success(t *testing.T) {
	const objectKey = "profiles/p/avatar-1.png"

	cases := []struct {
		name  string
		role  string
		table string
	}{
		{name: "patient", role: "PATIENT", table: "patient_profiles"},
		{name: "expert", role: "EXPERT", table: "expert_profiles"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mock := SetupTestDB(t)
			uploader := &fakeAvatarUploader{url: objectKey}
			h := NewAvatarHandler(uploader)

			authID := uuid.New()
			expectProfileLookup(mock, uuid.New(), authID, tc.role)
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE "` + tc.table + `" SET "avatar_url"`).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			body, ct := multipartAvatarBody(t, "file", pngBytes)
			c, w := newAvatarContext(t, authID, authID, tc.role, body, ct)

			h.UploadAvatar(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, 1, uploader.calls)

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assertBaseResponse(t, w, resp)
			result := resp["result"].(map[string]interface{})
			assert.Equal(t, objectKey, result["object_key"])
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
