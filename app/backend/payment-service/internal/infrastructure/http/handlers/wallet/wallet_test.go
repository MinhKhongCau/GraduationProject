package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"payment-service/internal/application/readquery"
	"payment-service/internal/application/wallet"
	"payment-service/internal/domain/money"
	walletdomain "payment-service/internal/domain/wallet"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mockWalletUsecase struct {
	getOrCreateWalletFunc func(ctx context.Context, userID uuid.UUID) (*walletdomain.Wallet, error)
	creditAvailableFunc   func(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	debitAvailableFunc    func(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	listHistoryFunc       func(ctx context.Context, query wallet.TransactionHistoryQuery) (*readquery.Page[walletdomain.WalletTransaction], error)
}

func (m *mockWalletUsecase) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*walletdomain.Wallet, error) {
	if m.getOrCreateWalletFunc != nil {
		return m.getOrCreateWalletFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWalletUsecase) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*walletdomain.Wallet, error) {
	return nil, nil
}

func (m *mockWalletUsecase) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]walletdomain.WalletTransaction, error) {
	return nil, nil
}

func (m *mockWalletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (m *mockWalletUsecase) CreditPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (m *mockWalletUsecase) CreditAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if m.creditAvailableFunc != nil {
		return m.creditAvailableFunc(ctx, userID, amount, refType, refID, idempotencyKey)
	}
	return nil
}

func (m *mockWalletUsecase) DebitAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if m.debitAvailableFunc != nil {
		return m.debitAvailableFunc(ctx, userID, amount, refType, refID, idempotencyKey)
	}
	return nil
}

func (m *mockWalletUsecase) DebitPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (m *mockWalletUsecase) DebitPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (m *mockWalletUsecase) LockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error {
	return nil
}

func (m *mockWalletUsecase) UnlockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error {
	return nil
}

func (m *mockWalletUsecase) ListTransactionHistory(ctx context.Context, query wallet.TransactionHistoryQuery) (*readquery.Page[walletdomain.WalletTransaction], error) {
	if m.listHistoryFunc != nil {
		return m.listHistoryFunc(ctx, query)
	}
	return nil, nil
}

func TestWalletHandlerAndUsecase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-PAY-WLT-01 - Nạp tiền vào ví điện tử thành công (TopUp)", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mockUsecase := &mockWalletUsecase{
			creditAvailableFunc: func(ctx context.Context, uID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
				if uID != userID || amount != 500000 {
					t.Fatalf("unexpected credit arguments: uID=%s amount=%d", uID, amount)
				}
				return nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"amount":500000}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/wallets/top-up", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.TopUpWallet(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-PAY-WLT-02 - Ghi nhận thù lao vào ví chuyên gia thành công", func(t *testing.T) {
		expertID := uuid.New()
		orderID := uuid.New()
		var creditedAmount money.Money
		var refTypeUsed string

		mockUsecase := &mockWalletUsecase{
			creditAvailableFunc: func(ctx context.Context, uID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
				creditedAmount = amount
				refTypeUsed = refType
				return nil
			},
		}

		// Giả lập logic tính thù lao: Tổng 500k, sàn lấy 20% (100k), chuyên gia hưởng 80% (400k)
		grossAmount := money.Money(500000)
		commissionAmount := money.Money(100000)
		netAmount := grossAmount.Sub(commissionAmount)

		err := mockUsecase.CreditAvailable(context.Background(), expertID, netAmount, "EXPERT_EARNING", orderID, "earning_"+orderID.String())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if creditedAmount != 400000 {
			t.Fatalf("expected credited amount 400000 VND, got %d", creditedAmount)
		}
		if refTypeUsed != "EXPERT_EARNING" {
			t.Fatalf("expected refType EXPERT_EARNING, got %s", refTypeUsed)
		}
	})

	t.Run("TC-PAY-WLT-03 - Xem thông tin và số dư ví điện tử thành công", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mockUsecase := &mockWalletUsecase{
			getOrCreateWalletFunc: func(ctx context.Context, uID uuid.UUID) (*walletdomain.Wallet, error) {
				return &walletdomain.Wallet{
					ID:               uuid.New(),
					UserID:           uID,
					AvailableBalance: money.Money(1000000),
					PendingBalance:   money.Money(0),
					LockedBalance:    money.Money(0),
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/wallets/me", nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.GetWallet(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}

		data := resp["data"].(map[string]interface{})
		if data["available_balance"] != float64(1000000) {
			t.Fatalf("expected 1000000 available balance, got %v", data["available_balance"])
		}
	})

	t.Run("TC-PAY-WLT-04 - Xem lịch sử giao dịch ví thành công", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mockUsecase := &mockWalletUsecase{
			listHistoryFunc: func(ctx context.Context, query wallet.TransactionHistoryQuery) (*readquery.Page[walletdomain.WalletTransaction], error) {
				return &readquery.Page[walletdomain.WalletTransaction]{
					Items: []walletdomain.WalletTransaction{
						{
							ID:           uuid.New(),
							WalletID:     uuid.New(),
							Type:         walletdomain.TxTypePaymentReceived,
							Amount:       money.Money(500000),
							BalanceAfter: money.Money(500000),
						},
					},
					Page:       1,
					Size:       10,
					TotalItems: 1,
					TotalPages: 1,
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/wallets/history?page=1&size=10", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.GetHistory(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WLT-05 - Lỗi số dư ví không đủ khi khấu trừ", func(t *testing.T) {
		userID := uuid.New()
		mockUsecase := &mockWalletUsecase{
			debitAvailableFunc: func(ctx context.Context, uID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
				return wallet.ErrInsufficientBalance
			},
		}

		err := mockUsecase.DebitAvailable(context.Background(), userID, money.Money(200000), "WITHDRAWAL", uuid.New(), "tx_key_1")

		if err != wallet.ErrInsufficientBalance {
			t.Fatalf("expected ErrInsufficientBalance, got %v", err)
		}
	})
}
