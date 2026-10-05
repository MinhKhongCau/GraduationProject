// Package httpclient cung cấp một HTTP client dùng chung cho toàn bộ payment-service
// khi cần gọi sang các service khác trong mạng nội bộ (internal call).
//
// Tính năng:
//   - Timeout mặc định 10 giây cho mỗi request.
//   - Retry tự động với exponential backoff (mặc định 3 lần) cho các lỗi mạng / 5xx.
//   - Inject Internal JWT token tự động qua TokenManager nếu được cung cấp.
//
// Cách dùng:
//
//	httpClient := httpclient.New(httpclient.Options{
//	    Timeout:    10 * time.Second,
//	    MaxRetries: 3,
//	})
//	resp, err := httpClient.Do(ctx, req)
package httpclient

import (
	"context"
	"log"
	"net/http"
	"time"
)

// TokenProvider là interface abstraction để inject token vào request.
// Được implement bởi *internalauth.TokenManager — tách biệt để không tạo circular import.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
	InvalidateToken()
}

// Options chứa cấu hình cho Client.
type Options struct {
	DisableRetries bool          // one request only; the caller owns retry policy
	Timeout        time.Duration // default: 10s
	MaxRetries     int           // default: 3; zero uses the default
	// TokenProvider nếu != nil sẽ tự động inject "Authorization: Bearer <token>" vào mỗi request.
	TokenProvider TokenProvider
}

// Client là HTTP client nội bộ.
type Client struct {
	httpClient    *http.Client
	maxRetries    int
	tokenProvider TokenProvider
}

// New tạo Client mới với các tùy chọn cho trước.
func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	maxRetries := opts.MaxRetries
	if opts.DisableRetries {
		maxRetries = 0
	} else if maxRetries == 0 {
		maxRetries = 3
	}

	return &Client{
		httpClient:    &http.Client{Timeout: timeout},
		maxRetries:    maxRetries,
		tokenProvider: opts.TokenProvider,
	}
}

// Do thực thi một HTTP request với retry + exponential backoff.
// Nếu có TokenProvider, token được inject vào header trước mỗi lần thử.
// Nếu nhận 401, token sẽ bị invalidate và retry ngay một lần.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Inject token nếu có TokenProvider
		if c.tokenProvider != nil {
			token, tokenErr := c.tokenProvider.GetToken(ctx)
			if tokenErr != nil {
				return nil, tokenErr
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			// Lỗi mạng → retry sau backoff
			if attempt < c.maxRetries {
				wait := backoff(attempt)
				log.Printf("[httpclient] attempt %d/%d failed (network error): %v — retrying in %s",
					attempt+1, c.maxRetries, err, wait)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(wait):
				}
			}
			continue
		}

		// Token hết hạn đột ngột (401): invalidate và retry 1 lần
		if resp.StatusCode == http.StatusUnauthorized && c.tokenProvider != nil && c.maxRetries == 0 {
			c.tokenProvider.InvalidateToken()
			return resp, nil
		}

		if resp.StatusCode == http.StatusUnauthorized && c.tokenProvider != nil {
			resp.Body.Close()
			c.tokenProvider.InvalidateToken()
			log.Printf("[httpclient] received 401, token invalidated — retrying request once")
			// Retry ngay, không cần backoff
			continue
		}

		// 5xx server error → retry với backoff
		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			resp.Body.Close()
			wait := backoff(attempt)
			log.Printf("[httpclient] attempt %d/%d: server returned %d — retrying in %s",
				attempt+1, c.maxRetries, resp.StatusCode, wait)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		// Request thành công hoặc lỗi 4xx (không retry)
		return resp, nil
	}

	return resp, err
}

// backoff tính thời gian chờ giữa các lần retry theo exponential backoff:
// attempt 0 → 200ms, attempt 1 → 400ms, attempt 2 → 800ms.
func backoff(attempt int) time.Duration {
	base := 200 * time.Millisecond
	for i := 0; i < attempt; i++ {
		base *= 2
	}
	return base
}
