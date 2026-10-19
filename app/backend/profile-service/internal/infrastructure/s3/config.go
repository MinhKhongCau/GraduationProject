// File: internal/infrastructure/s3/config.go
package s3

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the settings used to talk to the S3 bucket that stores profile images.
// The public URL of the bucket is not needed here: the service stores only object keys,
// and the frontend prefixes them with its own public URL (NEXT_PUBLIC_S3_PUBLIC_URL).
type Config struct {
	Bucket          string // S3_BUCKET
	Region          string // AWS_REGION ("auto" for Cloudflare R2; defaults to "auto" when Endpoint is set)
	Endpoint        string // S3_ENDPOINT (optional): S3-compatible API endpoint, e.g. https://<account_id>.r2.cloudflarestorage.com
	AccessKeyID     string // AWS_ACCESS_KEY_ID
	SecretAccessKey string // AWS_SECRET_ACCESS_KEY
}

// LoadConfig reads the S3 settings from environment variables and fails if any of them is missing.
func LoadConfig() (Config, error) {
	cfg := Config{
		Bucket:          os.Getenv("S3_BUCKET"),
		Region:          os.Getenv("AWS_REGION"),
		Endpoint:        strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	// Cloudflare R2 (and most S3-compatible stores) ignore the region; R2 expects "auto".
	if cfg.Region == "" && cfg.Endpoint != "" {
		cfg.Region = "auto"
	}

	var missing []string
	if cfg.Bucket == "" {
		missing = append(missing, "S3_BUCKET")
	}
	if cfg.Region == "" {
		missing = append(missing, "AWS_REGION")
	}
	if cfg.AccessKeyID == "" {
		missing = append(missing, "AWS_ACCESS_KEY_ID")
	}
	if cfg.SecretAccessKey == "" {
		missing = append(missing, "AWS_SECRET_ACCESS_KEY")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("thiếu biến môi trường S3: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
