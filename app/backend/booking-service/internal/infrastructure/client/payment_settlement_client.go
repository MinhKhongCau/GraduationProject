package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PaymentSettlementClient gọi payment-service để chi trả thù lao chuyên gia sau buổi tư vấn:
// POST {PAYMENT_SERVICE_INTERNAL_URL}/internal/payments/appointments/:id/settle (M2M JWT).
type PaymentSettlementClient struct {
	baseURL    string
	tokens     *TokenManager
	httpClient *http.Client
}

func NewPaymentSettlementClient(baseURL string, tokens *TokenManager) *PaymentSettlementClient {
	return &PaymentSettlementClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		tokens:     tokens,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *PaymentSettlementClient) SettleSession(ctx context.Context, appointmentID string) error {
	status, body, err := c.post(ctx, appointmentID)
	if err == nil && status == http.StatusUnauthorized {
		// Token có thể đã bị thu hồi/hết hạn sớm: lấy token mới và thử lại một lần.
		c.tokens.InvalidateToken()
		status, body, err = c.post(ctx, appointmentID)
	}
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("payment settlement returned %d: %s", status, body)
	}
	return nil
}

func (c *PaymentSettlementClient) post(ctx context.Context, appointmentID string) (int, string, error) {
	token, err := c.tokens.GetToken(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("payment settlement: get internal token: %w", err)
	}
	endpoint := c.baseURL + "/internal/payments/appointments/" + url.PathEscape(appointmentID) + "/settle"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("payment settlement: call %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, string(body), nil
}
