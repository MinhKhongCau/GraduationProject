// File: internal/infrastructure/s3/client_test.go
package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Smallest valid PNG header bytes, enough for http.DetectContentType.
var pngHeader = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0x0d}

func TestLoadConfig_MissingVariables(t *testing.T) {
	t.Setenv("S3_BUCKET", "")
	t.Setenv("S3_URL", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	_, err := LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "S3_BUCKET")
	assert.Contains(t, err.Error(), "AWS_SECRET_ACCESS_KEY")
}

func TestLoadConfig_TrimsTrailingSlashFromURL(t *testing.T) {
	t.Setenv("S3_BUCKET", "mindcare-profiles")
	t.Setenv("S3_URL", "https://cdn.example.com/")
	t.Setenv("AWS_REGION", "ap-southeast-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "key-id")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "mindcare-profiles", cfg.Bucket)
	assert.Equal(t, "https://cdn.example.com", cfg.URL)
}

func TestDetectImageType_AcceptsPNG(t *testing.T) {
	body := bytes.NewReader(pngHeader)

	contentType, err := detectImageType(body)
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)

	pos, _ := body.Seek(0, 1)
	assert.Equal(t, int64(0), pos, "reader must be rewound after detection")
}

func TestDetectImageType_RejectsNonImage(t *testing.T) {
	_, err := detectImageType(bytes.NewReader([]byte("plain text, not an image")))
	assert.True(t, errors.Is(err, ErrUnsupportedImageType))
}

func TestUploadProfileImage_RejectsNonImageBeforeTouchingS3(t *testing.T) {
	// The client has no real endpoint; the call must fail on validation before any network request.
	c := &Client{bucket: "b", baseURL: "https://cdn.example.com"}

	_, err := c.UploadProfileImage(context.Background(), "profile-1", bytes.NewReader([]byte("%PDF-1.4 ...")))
	assert.True(t, errors.Is(err, ErrUnsupportedImageType))
}

func TestObjectURL(t *testing.T) {
	c := &Client{baseURL: "https://cdn.example.com"}
	assert.Equal(t, "https://cdn.example.com/profiles/p1/a.png", c.ObjectURL("profiles/p1/a.png"))
}
