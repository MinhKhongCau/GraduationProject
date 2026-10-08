// Package profilegrpc gọi profile-service qua gRPC (mạng Docker nội bộ, không TLS).
package profilegrpc

import (
	"context"
	"fmt"
	"time"

	"payment-service/internal/application/managedscope"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/infrastructure/grpc/profilepb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const requestTimeout = 3 * time.Second

// Client triển khai managedscope.Resolver.
type Client struct {
	conn   *grpc.ClientConn
	client profilepb.ProfileQueryServiceClient
}

var _ managedscope.Resolver = (*Client)(nil)

// New tạo kết nối lazy: không lỗi khi profile-service chưa sẵn sàng lúc khởi động.
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("init profile gRPC client %s: %w", addr, err)
	}
	return &Client{conn: conn, client: profilepb.NewProfileQueryServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// ListManagedExpertIDs trả về auth id các chuyên gia do adminID quản lý.
func (c *Client) ListManagedExpertIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	resp, err := c.client.ListManagedExpertIds(ctx, &profilepb.ListManagedExpertIdsRequest{AdminId: adminID.String()})
	if err != nil {
		return nil, fmt.Errorf("profile-service ListManagedExpertIds: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(resp.GetExpertIds()))
	for _, raw := range resp.GetExpertIds() {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("profile-service returned invalid expert id %q: %w", raw, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

var _ apppayment.ProfileDirectory = (*Client)(nil)

// summaryBatch khớp giới hạn số id mỗi lần của profile-service.
const summaryBatch = 100

// GetProfileSummaries tra tên/ảnh/email của bệnh nhân và chuyên gia, tự chia lô 100 id.
func (c *Client) GetProfileSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]apppayment.PartySummary, error) {
	result := make(map[uuid.UUID]apppayment.PartySummary, len(ids))
	for start := 0; start < len(ids); start += summaryBatch {
		end := min(start+summaryBatch, len(ids))
		raw := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			raw = append(raw, id.String())
		}
		callCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		resp, err := c.client.GetProfileSummaries(callCtx, &profilepb.GetProfileSummariesRequest{AuthIds: raw})
		cancel()
		if err != nil {
			return result, fmt.Errorf("profile-service GetProfileSummaries: %w", err)
		}
		for _, p := range resp.GetProfiles() {
			id, err := uuid.Parse(p.GetAuthId())
			if err != nil {
				continue
			}
			result[id] = apppayment.PartySummary{ID: id, FullName: p.GetFullName(), AvatarURL: p.GetAvatarUrl(), Email: p.GetEmail()}
		}
	}
	return result, nil
}
