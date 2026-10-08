// Package profilegrpc gọi profile-service qua gRPC (mạng Docker nội bộ, không TLS).
package profilegrpc

import (
	"context"
	"fmt"
	"time"

	"payment-service/internal/application/managedscope"
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
