package withdrawal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"payment-service/internal/application/readquery"
	"payment-service/internal/application/wallet"
	"payment-service/internal/application/withdrawal"
	"payment-service/internal/domain/money"
	withdrawaldomain "payment-service/internal/domain/withdrawal"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockWithdrawalUsecase struct {
	linkBankFunc        func(ctx context.Context, userID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*withdrawaldomain.BankAccount, error)
	createWtdFunc       func(ctx context.Context, userID uuid.UUID, bankAccountID uuid.UUID, amount money.Money) (*withdrawaldomain.WithdrawalRequest, error)
	approveWtdFunc      func(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error
	rejectWtdFunc       func(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error
	getBankAccountsFunc func(ctx context.Context, userID uuid.UUID) ([]withdrawaldomain.BankAccount, error)
	listWithdrawalsFunc func(ctx context.Context, filter withdrawal.WithdrawalFilter) (*readquery.Page[withdrawal.WithdrawalView], error)
	getWithdrawalFunc   func(ctx context.Context, actorID, requestID uuid.UUID, isAdmin bool) (*withdrawal.WithdrawalView, error)
}

func (m *mockWithdrawalUsecase) LinkBankAccount(ctx context.Context, userID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*withdrawaldomain.BankAccount, error) {
	if m.linkBankFunc != nil {
		return m.linkBankFunc(ctx, userID, bankCode, accountNumber, accountHolderName)
	}
	return nil, nil
}

func (m *mockWithdrawalUsecase) GetBankAccounts(ctx context.Context, userID uuid.UUID) ([]withdrawaldomain.BankAccount, error) {
	if m.getBankAccountsFunc != nil {
		return m.getBankAccountsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWithdrawalUsecase) CreateWithdrawal(ctx context.Context, userID uuid.UUID, bankAccountID uuid.UUID, amount money.Money) (*withdrawaldomain.WithdrawalRequest, error) {
	if m.createWtdFunc != nil {
		return m.createWtdFunc(ctx, userID, bankAccountID, amount)
	}
	return nil, nil
}

func (m *mockWithdrawalUsecase) ApproveWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error {
	if m.approveWtdFunc != nil {
		return m.approveWtdFunc(ctx, adminID, requestID, note)
	}
	return nil
}

func (m *mockWithdrawalUsecase) RejectWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error {
	if m.rejectWtdFunc != nil {
		return m.rejectWtdFunc(ctx, adminID, requestID, note)
	}
	return nil
}

func (m *mockWithdrawalUsecase) HandlePayoutCallback(ctx context.Context, requestID uuid.UUID, payoutRef string, success bool, reason string) error {
	return nil
}

func (m *mockWithdrawalUsecase) ListWithdrawals(ctx context.Context, filter withdrawal.WithdrawalFilter) (*readquery.Page[withdrawal.WithdrawalView], error) {
	if m.listWithdrawalsFunc != nil {
		return m.listWithdrawalsFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockWithdrawalUsecase) GetWithdrawal(ctx context.Context, actorID, requestID uuid.UUID, isAdmin bool) (*withdrawal.WithdrawalView, error) {
	if m.getWithdrawalFunc != nil {
		return m.getWithdrawalFunc(ctx, actorID, requestID, isAdmin)
	}
	return nil, nil
}

func TestWithdrawalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-PAY-WTD-01 - Liên kết tài khoản ngân hàng thành công (Happy Case)", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mockUsecase := &mockWithdrawalUsecase{
			linkBankFunc: func(ctx context.Context, uID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*withdrawaldomain.BankAccount, error) {
				return &withdrawaldomain.BankAccount{
					ID:                uuid.New(),
					UserID:            uID,
					BankCode:          bankCode,
					AccountNumber:     accountNumber,
					AccountHolderName: accountHolderName,
					Verified:          true,
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{
			"bank_code": "MB",
			"account_number": "999999999",
			"account_holder_name": "NGUYEN VAN A"
		}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/bank-accounts", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.LinkBankAccount(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-PAY-WTD-02 - Liên kết ngân hàng thất bại do tên chủ thẻ không khớp", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mockUsecase := &mockWithdrawalUsecase{
			linkBankFunc: func(ctx context.Context, uID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*withdrawaldomain.BankAccount, error) {
				return nil, errors.New("tên chủ tài khoản không khớp với thông tin đăng ký tại ngân hàng MB")
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{
			"bank_code": "MB",
			"account_number": "1011223344",
			"account_holder_name": "SAI TEN CHU THE"
		}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/bank-accounts", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.LinkBankAccount(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-03 - Tạo phiếu yêu cầu rút tiền thành công (Happy Case)", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		bankAccID := uuid.New()

		mockUsecase := &mockWithdrawalUsecase{
			createWtdFunc: func(ctx context.Context, uID uuid.UUID, bAccID uuid.UUID, amount money.Money) (*withdrawaldomain.WithdrawalRequest, error) {
				return &withdrawaldomain.WithdrawalRequest{
					ID:            uuid.New(),
					WalletID:      uuid.New(),
					BankAccountID: bAccID,
					Amount:        amount,
					Status:        withdrawaldomain.WithdrawalStatusApproved,
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{
			"bank_account_id": "` + bankAccID.String() + `",
			"amount": 2000000
		}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/withdrawals", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.CreateWithdrawal(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-04 - Tạo phiếu rút tiền thất bại do số dư không đủ", func(t *testing.T) {
		userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		bankAccID := uuid.New()

		mockUsecase := &mockWithdrawalUsecase{
			createWtdFunc: func(ctx context.Context, uID uuid.UUID, bAccID uuid.UUID, amount money.Money) (*withdrawaldomain.WithdrawalRequest, error) {
				return nil, wallet.ErrInsufficientBalance
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{
			"bank_account_id": "` + bankAccID.String() + `",
			"amount": 50000000
		}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/withdrawals", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.CreateWithdrawal(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-05 - Admin phê duyệt phiếu giải ngân thù lao thành công", func(t *testing.T) {
		adminID := uuid.New()
		reqID := uuid.New()

		mockUsecase := &mockWithdrawalUsecase{
			approveWtdFunc: func(ctx context.Context, admID uuid.UUID, rID uuid.UUID, note string) error {
				return nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"note": "Đã chuyển khoản thành công"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/withdrawals/"+reqID.String()+"/approve", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "ADMIN")
		c.Request.Header.Set("X-User-Id", adminID.String())
		c.Params = gin.Params{{Key: "id", Value: reqID.String()}}

		h.ApproveWithdrawal(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-06 - Admin từ chối phiếu rút tiền thành công", func(t *testing.T) {
		adminID := uuid.New()
		reqID := uuid.New()

		mockUsecase := &mockWithdrawalUsecase{
			rejectWtdFunc: func(ctx context.Context, admID uuid.UUID, rID uuid.UUID, note string) error {
				return nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"note": "Thông tin tài khoản không hợp lệ"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/withdrawals/"+reqID.String()+"/reject", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "ADMIN")
		c.Request.Header.Set("X-User-Id", adminID.String())
		c.Params = gin.Params{{Key: "id", Value: reqID.String()}}

		h.RejectWithdrawal(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-07 - Admin phê duyệt thất bại do phiếu không ở trạng thái chờ duyệt", func(t *testing.T) {
		adminID := uuid.New()
		reqID := uuid.New()

		mockUsecase := &mockWithdrawalUsecase{
			approveWtdFunc: func(ctx context.Context, admID uuid.UUID, rID uuid.UUID, note string) error {
				return errors.New("yêu cầu rút tiền không ở trạng thái chờ duyệt")
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"note": "Approve lại"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/withdrawals/"+reqID.String()+"/approve", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "ADMIN")
		c.Request.Header.Set("X-User-Id", adminID.String())
		c.Params = gin.Params{{Key: "id", Value: reqID.String()}}

		h.ApproveWithdrawal(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-WTD-08 - Lấy danh sách tài khoản ngân hàng liên kết thành công", func(t *testing.T) {
		userID := uuid.New()
		mockUsecase := &mockWithdrawalUsecase{
			getBankAccountsFunc: func(ctx context.Context, uID uuid.UUID) ([]withdrawaldomain.BankAccount, error) {
				return []withdrawaldomain.BankAccount{
					{
						ID:            uuid.New(),
						UserID:        uID,
						BankCode:      "VCB",
						AccountNumber: "1234567890",
						Verified:      true,
					},
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/bank-accounts", nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.GetBankAccounts(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("TC-PAY-WTD-09 - Lấy danh sách phiếu rút tiền thành công (Expert)", func(t *testing.T) {
		userID := uuid.New()
		mockUsecase := &mockWithdrawalUsecase{
			listWithdrawalsFunc: func(ctx context.Context, filter withdrawal.WithdrawalFilter) (*readquery.Page[withdrawal.WithdrawalView], error) {
				return &readquery.Page[withdrawal.WithdrawalView]{
					Items:      []withdrawal.WithdrawalView{{ID: uuid.New(), AmountVND: 500000, Status: "APPROVED"}},
					TotalItems: 1,
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/withdrawals?page=1&size=10", nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())

		h.ListWithdrawals(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("TC-PAY-WTD-10 - Lấy danh sách phiếu rút tiền thành công (Admin)", func(t *testing.T) {
		adminID := uuid.New()
		mockUsecase := &mockWithdrawalUsecase{
			listWithdrawalsFunc: func(ctx context.Context, filter withdrawal.WithdrawalFilter) (*readquery.Page[withdrawal.WithdrawalView], error) {
				return &readquery.Page[withdrawal.WithdrawalView]{
					Items:      []withdrawal.WithdrawalView{{ID: uuid.New(), AmountVND: 1000000, Status: "PENDING"}},
					TotalItems: 1,
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/admin/withdrawals?page=1&size=10", nil)
		c.Request.Header.Set("X-User-Role", "ADMIN")
		c.Request.Header.Set("X-User-Id", adminID.String())

		h.ListAdminWithdrawals(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("TC-PAY-WTD-11 - Xem chi tiết phiếu rút tiền thành công", func(t *testing.T) {
		userID := uuid.New()
		wtdID := uuid.New()
		mockUsecase := &mockWithdrawalUsecase{
			getWithdrawalFunc: func(ctx context.Context, actorID, requestID uuid.UUID, isAdmin bool) (*withdrawal.WithdrawalView, error) {
				return &withdrawal.WithdrawalView{
					ID:        requestID,
					AmountVND: 200000,
					Status:    "PENDING",
				}, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/withdrawals/"+wtdID.String(), nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", userID.String())
		c.Params = gin.Params{{Key: "id", Value: wtdID.String()}}

		h.GetWithdrawal(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}
