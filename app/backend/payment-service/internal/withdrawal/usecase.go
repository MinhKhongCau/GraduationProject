package withdrawal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"payment-service/internal/payment/application/readquery"
	"payment-service/internal/wallet"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Usecase interface {
	LinkBankAccount(ctx context.Context, userID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*entity.BankAccount, error)
	GetBankAccounts(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error)

	CreateWithdrawal(ctx context.Context, userID uuid.UUID, bankAccountID uuid.UUID, amount vo.Money) (*entity.WithdrawalRequest, error)
	ApproveWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error
	RejectWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error
	HandlePayoutCallback(ctx context.Context, requestID uuid.UUID, payoutRef string, success bool, reason string) error
	ListWithdrawals(ctx context.Context, filter WithdrawalFilter) (*readquery.Page[WithdrawalView], error)
	GetWithdrawal(ctx context.Context, actorID, requestID uuid.UUID, isAdmin bool) (*WithdrawalView, error)
}

type withdrawalUsecase struct {
	repo          Repository
	walletUsecase wallet.Usecase
}

func NewUsecase(repo Repository, walletUsecase wallet.Usecase) Usecase {
	return &withdrawalUsecase{
		repo:          repo,
		walletUsecase: walletUsecase,
	}
}

func (u *withdrawalUsecase) LinkBankAccount(ctx context.Context, userID uuid.UUID, bankCode, accountNumber, accountHolderName string) (*entity.BankAccount, error) {
	// MOCK VALIDATION: Giả lập kiểm tra tài khoản ngân hàng từ VietQR/Napas
	// Nếu số tài khoản test là 1011223344 thì bắt buộc tên chủ thẻ phải là NGUYEN VAN B
	if accountNumber == "1011223344" && accountHolderName != "NGUYEN VAN B" {
		return nil, errors.New("tên chủ tài khoản không khớp với thông tin đăng ký tại ngân hàng " + bankCode)
	}

	account := &entity.BankAccount{
		ID:                uuid.New(),
		UserID:            userID,
		BankCode:          bankCode,
		AccountNumber:     accountNumber,
		AccountHolderName: accountHolderName,
		Verified:          true, // Mặc định verify trong dev
	}
	err := u.repo.CreateBankAccount(account)
	return account, err
}

func (u *withdrawalUsecase) GetBankAccounts(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error) {
	return u.repo.GetBankAccountsByUserID(userID)
}

func (u *withdrawalUsecase) CreateWithdrawal(ctx context.Context, userID uuid.UUID, bankAccountID uuid.UUID, amount vo.Money) (*entity.WithdrawalRequest, error) {
	// 1. Kiểm tra tài khoản ngân hàng
	bankAccount, err := u.repo.GetBankAccountByID(bankAccountID)
	if err != nil {
		return nil, err
	}
	if bankAccount.UserID != userID {
		return nil, errors.New("tài khoản ngân hàng không thuộc về chuyên gia này")
	}
	if !bankAccount.Verified {
		return nil, errors.New("tài khoản ngân hàng chưa được xác minh")
	}

	// 2. Kiểm tra số dư ví
	wlt, err := u.walletUsecase.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	if wlt.AvailableBalance.Int64() < amount.Int64() {
		return nil, wallet.ErrInsufficientBalance
	}

	// Ngưỡng rút tự động duyệt là 5,000,000 VND
	threshold := vo.Money(5000000)
	requiresManual := amount.Int64() > threshold.Int64()

	status := entity.WithdrawalStatusApproved
	if requiresManual {
		status = entity.WithdrawalStatusPendingApproval
	}

	req := &entity.WithdrawalRequest{
		ID:                     uuid.New(),
		WalletID:               wlt.ID,
		BankAccountID:          bankAccountID,
		Amount:                 amount,
		Status:                 status,
		RequiresManualApproval: requiresManual,
		RequestedAt:            time.Now().UnixMilli(),
	}

	err = u.repo.WithTransaction(func(tx *gorm.DB) error {
		// Khóa số tiền trong ví (chuyển available -> locked)
		err := u.walletUsecase.LockFunds(ctx, userID, amount)
		if err != nil {
			return err
		}

		// Tạo yêu cầu rút
		if err := u.repo.UpdateWithdrawalRequestWithTx(tx, req); err != nil {
			return err
		}

		// Ghi transaction log cho hành động khóa tiền (amount = 0 vì tổng số dư ví không đổi)
		transaction := &entity.WalletTransaction{
			ID:             uuid.New(),
			WalletID:       wlt.ID,
			Type:           entity.TxTypeWithdrawalLocked,
			Amount:         0,
			BalanceAfter:   wlt.AvailableBalance.Sub(amount).Add(wlt.PendingBalance).Add(wlt.LockedBalance.Add(amount)),
			ReferenceType:  "WITHDRAWAL_REQUEST",
			ReferenceID:    req.ID,
			IdempotencyKey: "lock_" + req.ID.String(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Tạo Outbox event
		eventType := "wallet.withdrawal.approved"
		if requiresManual {
			eventType = "wallet.withdrawal.pending_approval"
		}
		payloadMap := map[string]interface{}{
			"withdrawal_id":   req.ID,
			"wallet_id":       req.WalletID,
			"amount":          req.Amount.Int64(),
			"requires_manual": requiresManual,
			"requested_at":    req.RequestedAt,
		}
		payloadBytes, _ := json.Marshal(payloadMap)

		outbox := &entity.OutboxEvent{
			AggregateType: "WITHDRAWAL_REQUEST",
			AggregateID:   req.ID,
			EventType:     eventType,
			Payload:       string(payloadBytes),
			Published:     false,
		}
		return u.repo.SaveOutboxEvent(tx, outbox)
	})

	if err != nil {
		return nil, err
	}

	// Nếu không cần duyệt thủ công, kích hoạt payout gateway chạy ngầm
	if !requiresManual {
		go u.triggerPayoutGateway(req)
	}

	return req, nil
}

func (u *withdrawalUsecase) ApproveWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error {
	req, err := u.repo.GetWithdrawalRequestByID(requestID)
	if err != nil {
		return err
	}

	if req.Status != entity.WithdrawalStatusPendingApproval {
		return errors.New("yêu cầu rút tiền không ở trạng thái chờ duyệt")
	}

	now := time.Now().UnixMilli()
	req.Status = entity.WithdrawalStatusApproved
	req.ApproverID = &adminID
	req.ApprovalAction = entity.ApprovalActionApprove
	req.ApprovalNote = note
	req.ApprovedAt = &now

	err = u.repo.WithTransaction(func(tx *gorm.DB) error {
		if err := u.repo.UpdateWithdrawalRequestWithTx(tx, req); err != nil {
			return err
		}

		payloadMap := map[string]interface{}{
			"withdrawal_id": req.ID,
			"wallet_id":     req.WalletID,
			"amount":        req.Amount.Int64(),
			"approved_at":   now,
			"approver_id":   adminID,
		}
		payloadBytes, _ := json.Marshal(payloadMap)

		outbox := &entity.OutboxEvent{
			AggregateType: "WITHDRAWAL_REQUEST",
			AggregateID:   req.ID,
			EventType:     "wallet.withdrawal.approved",
			Payload:       string(payloadBytes),
			Published:     false,
		}
		return u.repo.SaveOutboxEvent(tx, outbox)
	})

	if err != nil {
		return err
	}

	// Kích hoạt payout gateway chạy ngầm
	go u.triggerPayoutGateway(req)

	return nil
}

func (u *withdrawalUsecase) RejectWithdrawal(ctx context.Context, adminID uuid.UUID, requestID uuid.UUID, note string) error {
	req, err := u.repo.GetWithdrawalRequestByID(requestID)
	if err != nil {
		return err
	}

	if req.Status != entity.WithdrawalStatusPendingApproval {
		return errors.New("yêu cầu rút tiền không ở trạng thái chờ duyệt")
	}

	now := time.Now().UnixMilli()
	req.Status = entity.WithdrawalStatusRejected
	req.ApproverID = &adminID
	req.ApprovalAction = entity.ApprovalActionReject
	req.ApprovalNote = note
	req.ProcessedAt = &now

	var walletObj entity.Wallet
	err = u.repo.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", req.WalletID).First(&walletObj).Error; err != nil {
			return err
		}

		// Rollback tiền ví: giải phóng locked -> available
		err := u.walletUsecase.UnlockFunds(ctx, walletObj.UserID, req.Amount)
		if err != nil {
			return err
		}

		// Cập nhật trạng thái phiếu
		if err := u.repo.UpdateWithdrawalRequestWithTx(tx, req); err != nil {
			return err
		}

		// Ghi transaction log (amount = 0 vì tổng ví không đổi)
		transaction := &entity.WalletTransaction{
			ID:             uuid.New(),
			WalletID:       req.WalletID,
			Type:           entity.TxTypeWithdrawalRejected,
			Amount:         0,
			BalanceAfter:   walletObj.AvailableBalance.Add(req.Amount).Add(walletObj.PendingBalance).Add(walletObj.LockedBalance.Sub(req.Amount)),
			ReferenceType:  "WITHDRAWAL_REQUEST",
			ReferenceID:    req.ID,
			IdempotencyKey: "reject_" + req.ID.String(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		payloadMap := map[string]interface{}{
			"withdrawal_id": req.ID,
			"wallet_id":     req.WalletID,
			"amount":        req.Amount.Int64(),
			"rejected_at":   now,
			"approver_id":   adminID,
			"note":          note,
		}
		payloadBytes, _ := json.Marshal(payloadMap)

		outbox := &entity.OutboxEvent{
			AggregateType: "WITHDRAWAL_REQUEST",
			AggregateID:   req.ID,
			EventType:     "wallet.withdrawal.rejected",
			Payload:       string(payloadBytes),
			Published:     false,
		}
		return u.repo.SaveOutboxEvent(tx, outbox)
	})

	return err
}

func (u *withdrawalUsecase) HandlePayoutCallback(ctx context.Context, requestID uuid.UUID, payoutRef string, success bool, reason string) error {
	req, err := u.repo.GetWithdrawalRequestByID(requestID)
	if err != nil {
		return err
	}

	if req.Status != entity.WithdrawalStatusProcessing {
		log.Printf("Yêu cầu rút %s đã được xử lý từ trước (status=%s)", requestID, req.Status)
		return nil
	}

	now := time.Now().UnixMilli()
	req.ProcessedAt = &now
	req.PayoutRef = payoutRef

	var walletObj entity.Wallet
	txErr := u.repo.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", req.WalletID).First(&walletObj).Error; err != nil {
			return err
		}

		if success {
			req.Status = entity.WithdrawalStatusCompleted

			// Khóa dòng ví và trừ tiền khỏi locked_balance (lúc này tổng ví thực sự giảm)
			var wlt entity.Wallet
			if err := tx.Clauses(gorm.Expr("FOR UPDATE")).Where("id = ?", req.WalletID).First(&wlt).Error; err != nil {
				return err
			}
			wlt.LockedBalance = wlt.LockedBalance.Sub(req.Amount)
			wlt.Version++
			if err := tx.Save(&wlt).Error; err != nil {
				return err
			}

			if err := u.repo.UpdateWithdrawalRequestWithTx(tx, req); err != nil {
				return err
			}

			// Ghi transaction log cho withdrawal completed (số tiền giảm -> âm!)
			transaction := &entity.WalletTransaction{
				ID:             uuid.New(),
				WalletID:       req.WalletID,
				Type:           entity.TxTypeWithdrawalCompleted,
				Amount:         vo.Money(-req.Amount.Int64()),
				BalanceAfter:   wlt.AvailableBalance.Add(wlt.PendingBalance).Add(wlt.LockedBalance),
				ReferenceType:  "WITHDRAWAL_REQUEST",
				ReferenceID:    req.ID,
				IdempotencyKey: "complete_" + req.ID.String(),
			}
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}

			payloadMap := map[string]interface{}{
				"withdrawal_id": req.ID,
				"wallet_id":     req.WalletID,
				"amount":        req.Amount.Int64(),
				"completed_at":  now,
				"payout_ref":    payoutRef,
			}
			payloadBytes, _ := json.Marshal(payloadMap)

			outbox := &entity.OutboxEvent{
				AggregateType: "WITHDRAWAL_REQUEST",
				AggregateID:   req.ID,
				EventType:     "wallet.withdrawal.completed",
				Payload:       string(payloadBytes),
				Published:     false,
			}
			return u.repo.SaveOutboxEvent(tx, outbox)

		} else {
			req.Status = entity.WithdrawalStatusFailed

			// Thất bại: rollback locked -> available
			err := u.walletUsecase.UnlockFunds(ctx, walletObj.UserID, req.Amount)
			if err != nil {
				return err
			}

			if err := u.repo.UpdateWithdrawalRequestWithTx(tx, req); err != nil {
				return err
			}

			// Ghi transaction log (amount = 0 vì tổng ví không đổi)
			transaction := &entity.WalletTransaction{
				ID:             uuid.New(),
				WalletID:       req.WalletID,
				Type:           entity.TxTypeWithdrawalRejected,
				Amount:         0,
				BalanceAfter:   walletObj.AvailableBalance.Add(req.Amount).Add(walletObj.PendingBalance).Add(walletObj.LockedBalance.Sub(req.Amount)),
				ReferenceType:  "WITHDRAWAL_REQUEST",
				ReferenceID:    req.ID,
				IdempotencyKey: "fail_" + req.ID.String(),
			}
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}

			payloadMap := map[string]interface{}{
				"withdrawal_id": req.ID,
				"wallet_id":     req.WalletID,
				"amount":        req.Amount.Int64(),
				"failed_at":     now,
				"reason":        reason,
			}
			payloadBytes, _ := json.Marshal(payloadMap)

			outbox := &entity.OutboxEvent{
				AggregateType: "WITHDRAWAL_REQUEST",
				AggregateID:   req.ID,
				EventType:     "wallet.withdrawal.failed",
				Payload:       string(payloadBytes),
				Published:     false,
			}
			return u.repo.SaveOutboxEvent(tx, outbox)
		}
	})

	return txErr
}

// Giả lập payout gateway gọi webhook callback async sau 2s
func (u *withdrawalUsecase) triggerPayoutGateway(req *entity.WithdrawalRequest) {
	log.Printf("[PAYOUT GATEWAY] Gọi cổng thanh toán cho yêu cầu ID %s", req.ID)

	// Đổi trạng thái sang PROCESSING
	err := u.repo.WithTransaction(func(tx *gorm.DB) error {
		req.Status = entity.WithdrawalStatusProcessing
		return u.repo.UpdateWithdrawalRequestWithTx(tx, req)
	})
	if err != nil {
		log.Printf("[PAYOUT GATEWAY ERROR] Đổi trạng thái PROCESSING thất bại: %v", err)
		return
	}

	time.AfterFunc(2*time.Second, func() {
		log.Printf("[PAYOUT GATEWAY] Callback webhook cho yêu cầu ID %s", req.ID)
		ctx := context.Background()

		success := true
		reason := ""
		payoutRef := "MOCK_PAYOUT_" + uuid.New().String()[:8]

		// Để test: Nếu số tiền kết thúc bằng số lẻ 99 (VND), giả lập lỗi
		if req.Amount.Int64()%100 == 99 {
			success = false
			reason = "Lỗi đường truyền hệ thống ngân hàng đối tác"
		}

		err = u.HandlePayoutCallback(ctx, req.ID, payoutRef, success, reason)
		if err != nil {
			log.Printf("[PAYOUT GATEWAY WEBHOOK ERROR] Callback thất bại cho ID %s: %v", req.ID, err)
		}
	})
}
