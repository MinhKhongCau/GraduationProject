package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	bookingclient "payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"payment-service/internal/payment/gateway"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestProcessIPNConcurrentRequestsOnlyProcessPendingOrderOnce(t *testing.T) {
	orderID := uuid.New()
	payerID := uuid.New()
	expertID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          payerID,
		ExpertID:         expertID,
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
	})

	usecase := NewUsecase(
		repo,
		&fakeWalletUsecase{delay: 50 * time.Millisecond},
		gateway.NewVNPayClient("", "", "", ""),
		&fakeBookingClient{},
	)

	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"00"},
		"vnp_TransactionNo": {"vnp-txn-1"},
	}
	signVNPayParams(params)

	start := make(chan struct{})
	type result struct {
		alreadyProcessed bool
		err              error
	}
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start
			alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
			results <- result{alreadyProcessed: alreadyProcessed, err: err}
		}()
	}
	close(start)

	var alreadyProcessedCount int
	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			t.Fatalf("ProcessIPN returned error: %v", res.err)
		}
		if res.alreadyProcessed {
			alreadyProcessedCount++
		}
	}

	if alreadyProcessedCount != 1 {
		t.Fatalf("expected exactly one alreadyProcessed response, got %d", alreadyProcessedCount)
	}
	if repo.successTransitions != 1 {
		t.Fatalf("expected exactly one PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if repo.order.Status != entity.OrderStatusSuccess {
		t.Fatalf("expected final order status SUCCESS, got %s", repo.order.Status.String())
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no outbox events for order without appointment, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNCommitsPaymentWalletAndBookingOutboxAtomically(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, walletUsecase, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.order.Status != entity.OrderStatusSuccess {
		t.Fatalf("expected order status SUCCESS, got %s", repo.order.Status.String())
	}
	if walletUsecase.pendingBalance != vo.Money(850) {
		t.Fatalf("expected wallet pending balance 850, got %d", walletUsecase.pendingBalance)
	}
	if len(walletUsecase.transactions) != 2 {
		t.Fatalf("expected two wallet transactions, got %d", len(walletUsecase.transactions))
	}
	assertOnlyBookingConfirmOutbox(t, repo)
}

func TestProcessIPNRollsBackWhenWalletCreditFails(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{creditErr: errors.New("credit failed")}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, walletUsecase, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err == nil {
		t.Fatal("expected wallet credit error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	assertRolledBackPaymentWalletAndOutbox(t, repo, walletUsecase)
}

func TestProcessIPNRollsBackCreditWhenWalletDebitFails(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{debitErr: errors.New("debit failed")}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, walletUsecase, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err == nil {
		t.Fatal("expected wallet debit error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	assertRolledBackPaymentWalletAndOutbox(t, repo, walletUsecase)
}

func TestProcessIPNAbortsWhenGatewayTxnLookupReturnsDatabaseError(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
	})
	repo.gatewayLookupErr = errors.New("database unavailable")

	usecase := NewUsecase(
		repo,
		&fakeWalletUsecase{},
		gateway.NewVNPayClient("", "", "", ""),
		&fakeBookingClient{},
	)

	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"00"},
		"vnp_TransactionNo": {"vnp-txn-1"},
	}
	signVNPayParams(params)
	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
	if err == nil {
		t.Fatal("expected database lookup error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no outbox events, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNRejectsMissingSecureHash(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
	})

	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, err := usecase.ProcessIPN(context.Background(), map[string][]string{
		"vnp_TxnRef":       {orderID.String()},
		"vnp_Amount":       {"100000"},
		"vnp_ResponseCode": {"00"},
	})
	if err == nil || !strings.Contains(err.Error(), "vnp_SecureHash") {
		t.Fatalf("expected missing secure hash error, got %v", err)
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
}

func TestProcessIPNRejectsInvalidSecureHash(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
	})

	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, err := usecase.ProcessIPN(context.Background(), map[string][]string{
		"vnp_TxnRef":       {orderID.String()},
		"vnp_Amount":       {"100000"},
		"vnp_ResponseCode": {"00"},
		"vnp_SecureHash":   {"invalid"},
	})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("expected invalid checksum error, got %v", err)
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
}

func TestProcessIPNRejectsMalformedAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount *string
	}{
		{name: "missing amount", amount: nil},
		{name: "invalid integer", amount: stringPtr("not-a-number")},
		{name: "zero amount", amount: stringPtr("0")},
		{name: "negative amount", amount: stringPtr("-100")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderID := uuid.New()
			repo := newFakePaymentRepo(entity.PaymentOrder{
				ID:               orderID,
				PayerID:          uuid.New(),
				ExpertID:         uuid.New(),
				GrossAmount:      vo.Money(1000),
				CommissionRate:   0.15,
				CommissionAmount: vo.Money(150),
				NetAmount:        vo.Money(850),
				Gateway:          "VNPAY",
				Status:           entity.OrderStatusPending,
			})

			usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})
			params := map[string][]string{
				"vnp_TxnRef":       {orderID.String()},
				"vnp_ResponseCode": {"00"},
			}
			if tt.amount != nil {
				params["vnp_Amount"] = []string{*tt.amount}
			}
			signVNPayParams(params)

			_, err := usecase.ProcessIPN(context.Background(), params)
			if err == nil || !strings.Contains(err.Error(), "vnp_Amount") {
				t.Fatalf("expected amount error, got %v", err)
			}
			if repo.successTransitions != 0 {
				t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
			}
		})
	}
}

func TestProcessIPNFailedPaymentWithAppointmentCreatesBookingFailOutbox(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(orderID))
	if err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.order.Status != entity.OrderStatusFailed {
		t.Fatalf("expected order FAILED, got %s", repo.order.Status.String())
	}
	assertOnlyBookingFailOutbox(t, repo)
}

func TestProcessIPNFailedPaymentWithoutAppointmentCreatesNoBookingEvent(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
	})
	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	if _, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(orderID)); err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no booking event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNInvalidChecksumCreatesNoFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	params := signedFailedIPNParams(orderID)
	params["vnp_SecureHash"] = []string{"invalid"}
	if _, err := usecase.ProcessIPN(context.Background(), params); err == nil {
		t.Fatal("expected invalid checksum error")
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no failure event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNAmountMismatchCreatesNoFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	params := signedFailedIPNParams(orderID)
	params["vnp_Amount"] = []string{"99900"}
	signVNPayParams(params)
	if _, err := usecase.ProcessIPN(context.Background(), params); err == nil {
		t.Fatal("expected amount mismatch error")
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no failure event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNDuplicateFailedPaymentDoesNotCreateDuplicateFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(entity.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      vo.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: vo.Money(150),
		NetAmount:        vo.Money(850),
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, &fakeWalletUsecase{}, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})
	params := signedFailedIPNParams(orderID)

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
	if err != nil {
		t.Fatalf("first ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected first IPN to process")
	}
	alreadyProcessed, err = usecase.ProcessIPN(context.Background(), params)
	if err != nil {
		t.Fatalf("second ProcessIPN returned error: %v", err)
	}
	if !alreadyProcessed {
		t.Fatal("expected duplicate IPN to be already processed")
	}
	assertOnlyBookingFailOutbox(t, repo)
}

type fakePaymentRepo struct {
	mu                 sync.Mutex
	rowLock            sync.Mutex
	rowLockHeld        bool
	order              entity.PaymentOrder
	outboxEvents       []entity.OutboxEvent
	successTransitions int
	gatewayLookupErr   error
	participants       []fakeTxParticipant
}

func newFakePaymentRepo(order entity.PaymentOrder) *fakePaymentRepo {
	return &fakePaymentRepo{order: order}
}

type fakeTxParticipant interface {
	beginTx()
	commitTx()
	rollbackTx()
}

func (r *fakePaymentRepo) addParticipant(participant fakeTxParticipant) {
	r.participants = append(r.participants, participant)
}

func (r *fakePaymentRepo) Create(order *entity.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = *order
	return nil
}

func (r *fakePaymentRepo) GetByID(orderID uuid.UUID) (*entity.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.ID != orderID {
		return nil, gorm.ErrRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetByIDForUpdate(tx *gorm.DB, orderID uuid.UUID) (*entity.PaymentOrder, error) {
	r.rowLock.Lock()
	r.mu.Lock()
	r.rowLockHeld = true
	defer r.mu.Unlock()

	if r.order.ID != orderID {
		return nil, gorm.ErrRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error) {
	return r.GetByGatewayTxnRefWithTx(nil, ref)
}

func (r *fakePaymentRepo) GetByGatewayTxnRefWithTx(tx *gorm.DB, ref string) (*entity.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.gatewayLookupErr != nil {
		return nil, r.gatewayLookupErr
	}
	if r.order.GatewayTxnRef != ref {
		return nil, gorm.ErrRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) Update(order *entity.PaymentOrder) error {
	return r.UpdateWithTx(nil, order)
}

func (r *fakePaymentRepo) UpdateWithTx(tx *gorm.DB, order *entity.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.Status == entity.OrderStatusPending && order.Status == entity.OrderStatusSuccess {
		r.successTransitions++
	}
	r.order = *order
	return nil
}

func (r *fakePaymentRepo) SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outboxEvents = append(r.outboxEvents, *event)
	return nil
}

func (r *fakePaymentRepo) WithTransaction(fn func(tx *gorm.DB) error) error {
	r.mu.Lock()
	orderSnapshot := r.order
	outboxSnapshot := append([]entity.OutboxEvent(nil), r.outboxEvents...)
	successTransitionsSnapshot := r.successTransitions
	for _, participant := range r.participants {
		participant.beginTx()
	}
	r.mu.Unlock()

	err := fn(nil)

	r.mu.Lock()
	if err != nil {
		r.order = orderSnapshot
		r.outboxEvents = outboxSnapshot
		r.successTransitions = successTransitionsSnapshot
		for _, participant := range r.participants {
			participant.rollbackTx()
		}
	} else {
		for _, participant := range r.participants {
			participant.commitTx()
		}
	}
	held := r.rowLockHeld
	if held {
		r.rowLockHeld = false
	}
	r.mu.Unlock()
	if held {
		r.rowLock.Unlock()
	}
	return err
}

type fakeWalletUsecase struct {
	delay                time.Duration
	creditErr            error
	debitErr             error
	pendingBalance       vo.Money
	transactions         []entity.WalletTransaction
	pendingBalanceBefore vo.Money
	transactionsBefore   []entity.WalletTransaction
}

func (u *fakeWalletUsecase) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	return &entity.Wallet{ID: uuid.New(), UserID: userID}, nil
}

func (u *fakeWalletUsecase) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*entity.Wallet, error) {
	return &entity.Wallet{ID: walletID}, nil
}

func (u *fakeWalletUsecase) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]entity.WalletTransaction, error) {
	return nil, nil
}

func (u *fakeWalletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.CreditPendingWithTx(ctx, nil, userID, amount, refType, refID, idempotencyKey)
}

func (u *fakeWalletUsecase) CreditPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if u.delay > 0 {
		time.Sleep(u.delay)
	}
	if u.creditErr != nil {
		return u.creditErr
	}
	u.pendingBalance = u.pendingBalance.Add(amount)
	u.transactions = append(u.transactions, entity.WalletTransaction{
		Type:           entity.TxTypePaymentReceived,
		Amount:         amount,
		ReferenceType:  refType,
		ReferenceID:    refID,
		IdempotencyKey: idempotencyKey,
	})
	return nil
}

func (u *fakeWalletUsecase) CreditAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (u *fakeWalletUsecase) DebitAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (u *fakeWalletUsecase) DebitPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.DebitPendingWithTx(ctx, nil, userID, amount, refType, refID, idempotencyKey)
}

func (u *fakeWalletUsecase) DebitPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if u.debitErr != nil {
		return u.debitErr
	}
	u.pendingBalance = u.pendingBalance.Sub(amount)
	u.transactions = append(u.transactions, entity.WalletTransaction{
		Type:           entity.TxTypeCommissionDeducted,
		Amount:         vo.Money(-amount.Int64()),
		ReferenceType:  refType,
		ReferenceID:    refID,
		IdempotencyKey: idempotencyKey,
	})
	return nil
}

func (u *fakeWalletUsecase) LockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error {
	return nil
}

func (u *fakeWalletUsecase) UnlockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error {
	return nil
}

func (u *fakeWalletUsecase) beginTx() {
	u.pendingBalanceBefore = u.pendingBalance
	u.transactionsBefore = append([]entity.WalletTransaction(nil), u.transactions...)
}

func (u *fakeWalletUsecase) commitTx() {
	u.transactionsBefore = nil
}

func (u *fakeWalletUsecase) rollbackTx() {
	u.pendingBalance = u.pendingBalanceBefore
	u.transactions = append([]entity.WalletTransaction(nil), u.transactionsBefore...)
	u.transactionsBefore = nil
}

type fakeBookingClient struct{}

func (c *fakeBookingClient) GetAppointment(ctx context.Context, appointmentID string) (*bookingclient.Appointment, error) {
	return &bookingclient.Appointment{AppointmentID: appointmentID}, nil
}

func (c *fakeBookingClient) ConfirmAppointment(ctx context.Context, appointmentID string) error {
	return nil
}

func (c *fakeBookingClient) FailAppointment(ctx context.Context, appointmentID string) error {
	return nil
}

func signedSuccessIPNParams(orderID uuid.UUID) map[string][]string {
	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"00"},
		"vnp_TransactionNo": {"vnp-txn-1"},
	}
	signVNPayParams(params)
	return params
}

func signedFailedIPNParams(orderID uuid.UUID) map[string][]string {
	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"24"},
		"vnp_TransactionNo": {"vnp-txn-failed-1"},
	}
	signVNPayParams(params)
	return params
}

func assertOnlyBookingConfirmOutbox(t *testing.T, repo *fakePaymentRepo) {
	t.Helper()

	if len(repo.outboxEvents) != 1 {
		t.Fatalf("expected exactly one outbox event, got %d", len(repo.outboxEvents))
	}
	if repo.outboxEvents[0].EventType != "booking.appointment.confirm" {
		t.Fatalf("expected booking.appointment.confirm outbox event, got %q", repo.outboxEvents[0].EventType)
	}
	for _, event := range repo.outboxEvents {
		if event.EventType == "wallet.payment.received" {
			t.Fatal("did not expect wallet.payment.received outbox event")
		}
	}
}

func assertOnlyBookingFailOutbox(t *testing.T, repo *fakePaymentRepo) {
	t.Helper()

	if len(repo.outboxEvents) != 1 {
		t.Fatalf("expected exactly one outbox event, got %d", len(repo.outboxEvents))
	}
	if repo.outboxEvents[0].EventType != "booking.appointment.fail" {
		t.Fatalf("expected booking.appointment.fail outbox event, got %q", repo.outboxEvents[0].EventType)
	}
}

func assertRolledBackPaymentWalletAndOutbox(t *testing.T, repo *fakePaymentRepo, walletUsecase *fakeWalletUsecase) {
	t.Helper()

	if repo.order.Status != entity.OrderStatusPending {
		t.Fatalf("expected order status to roll back to PENDING, got %s", repo.order.Status.String())
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no committed PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected outbox events to roll back, got %d", len(repo.outboxEvents))
	}
	if walletUsecase.pendingBalance != 0 {
		t.Fatalf("expected wallet pending balance to roll back to 0, got %d", walletUsecase.pendingBalance)
	}
	if len(walletUsecase.transactions) != 0 {
		t.Fatalf("expected wallet transactions to roll back, got %d", len(walletUsecase.transactions))
	}
}

func signVNPayParams(params map[string][]string) {
	cleanParams := make(map[string]string)
	for k, vals := range params {
		if len(vals) == 0 {
			continue
		}
		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" {
			continue
		}
		cleanParams[k] = vals[0]
	}

	keys := make([]string, 0, len(cleanParams))
	for k := range cleanParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(cleanParams[k])))
	}

	mac := hmac.New(sha512.New, []byte("MOCK_SECRET"))
	mac.Write([]byte(queryBuilder.String()))
	params["vnp_SecureHash"] = []string{hex.EncodeToString(mac.Sum(nil))}
}

func stringPtr(s string) *string {
	return &s
}
