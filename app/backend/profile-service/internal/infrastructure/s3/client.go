// File: internal/infrastructure/s3/client.go
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// ErrUnsupportedImageType is returned when the uploaded file is not a JPEG, PNG, GIF or WebP image.
var ErrUnsupportedImageType = errors.New("định dạng ảnh không được hỗ trợ (chỉ JPEG, PNG, GIF, WebP)")

// allowedImageExtensions maps the content type detected from the file bytes to the file extension.
var allowedImageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// Client uploads objects to a single S3 bucket.
type Client struct {
	api    *awss3.Client
	bucket string
}

// NewClient creates an S3 client with the static credentials from cfg.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo cấu hình AWS: %w", err)
	}

	api := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if cfg.Endpoint == "" {
			return // Amazon S3: the SDK resolves the endpoint from the region.
		}
		// S3-compatible store such as Cloudflare R2.
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
		// Only send/verify checksums when the API requires them; newer SDK defaults
		// add checksum headers that S3-compatible stores may reject.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &Client{
		api:    api,
		bucket: cfg.Bucket,
	}, nil
}

// UploadProfileImage checks that body is an image, stores it under profiles/<profileID>/ with a
// unique name, and returns the object key (e.g. profiles/<profileID>/avatar-<uuid>.png).
// Callers store the key; clients build the public URL by prefixing it with the bucket's public URL.
// The content type is detected from the file bytes, not from the client-supplied header.
func (c *Client) UploadProfileImage(ctx context.Context, profileID string, body io.ReadSeeker) (string, error) {
	contentType, err := detectImageType(body)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("profiles/%s/avatar-%s%s", profileID, uuid.NewString(), allowedImageExtensions[contentType])
	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("không thể đọc lại file ảnh: %w", err)
	}

	_, err = c.api.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload ảnh lên S3 thất bại: %w", err)
	}

	return key, nil
}

// detectImageType sniffs the first 512 bytes of body and returns its image content type.
// It rewinds body to the start before returning.
func detectImageType(body io.ReadSeeker) (string, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(body, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("không thể đọc file ảnh: %w", err)
	}
	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("không thể đọc lại file ảnh: %w", err)
	}

	contentType := http.DetectContentType(head[:n])
	if _, ok := allowedImageExtensions[contentType]; !ok {
		return "", ErrUnsupportedImageType
	}
	return contentType, nil
}
