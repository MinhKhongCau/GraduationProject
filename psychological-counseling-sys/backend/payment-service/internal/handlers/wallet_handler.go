// File: internal/handlers/wallet_handler.go
package handlers

import (
	"errors"
	"net/http"
	"payment-service/config"
	"payment-service/internal/models"
	"payment-service/internal/schemas"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 1. KHỞI TẠO VÍ (0đ)
func InitWallet(c *gin.Context) {
	var req schemas.InitWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	ownerUUID, _ := uuid.Parse(req.OwnerID)
	wallet := models.Wallet{
		OwnerID:  ownerUUID,
		UserType: req.UserType,
		// Balance mặc định là 0 do GORM tự set
	}

	if err := config.DB.Create(&wallet).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Ví của người dùng này đã tồn tại!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Khởi tạo ví thành công", "data": wallet})
}

// 2. XEM SỐ DƯ VÍ
func GetWallet(c *gin.Context) {
	ownerID := c.Param("owner_id")
	var wallet models.Wallet

	if err := config.DB.Where("owner_id = ?", ownerID).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy ví!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": wallet})
}

// 3. NẠP TIỀN (TOP-UP) VỚI GORM TRANSACTION & ROW LOCKING
func TopUpWallet(c *gin.Context) {
	ownerID := c.Param("owner_id")
	var req schemas.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.NewFromInt(0)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền nạp phải lớn hơn 0"})
		return
	}

	// BẮT ĐẦU TRANSACTION: Nếu có lỗi ở bất kỳ bước nào, toàn bộ quá trình sẽ bị Hủy bỏ (Rollback)
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		// BƯỚC A: Tìm Ví và KHÓA DÒNG (SELECT ... FOR UPDATE)
		// Khóa này ngăn không cho ai khác chạm vào ví này cho đến khi giao dịch nạp tiền này kết thúc.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", ownerID).First(&wallet).Error; err != nil {
			return err // Báo lỗi để Rollback
		}

		balanceBefore := wallet.Balance
		balanceAfter := balanceBefore.Add(req.Amount) // Cộng tiền bằng thư viện decimal

		// BƯỚC B: Cập nhật số dư mới vào Ví
		wallet.Balance = balanceAfter
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		// BƯỚC C: Ghi lại Lịch sử giao dịch
		// Dùng UUID giả cho RelatedOrderID trong lúc nạp tiền tự do
		transaction := models.Transaction{
			WalletID:        wallet.WalletID,
			RelatedOrderID:  uuid.New(),
			TransactionType: "TOP_UP",
			Amount:          req.Amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			Status:          "SUCCESS",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Nếu chạy đến đây mà không có lỗi (return err) nào -> Sẽ tự động COMMIT lưu vào DB
		return nil
	})

	// Kiểm tra kết quả của Transaction
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Giao dịch thất bại: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nạp tiền thành công!"})
}

// 4. THANH TOÁN DỊCH VỤ (Trừ tiền Patient, Cộng tiền Expert)
func ProcessPayment(c *gin.Context) {
	var req schemas.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Tỷ lệ hoa hồng (Ví dụ: 15%)
	commissionRate := decimal.NewFromFloat(0.15)

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var payerWallet, expertWallet, systemWallet models.Wallet

		// 1. Khóa và trừ tiền Ví Bệnh nhân (100% số tiền)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", req.PayerID).First(&payerWallet).Error; err != nil {
			return err
		}
		if payerWallet.Balance.LessThan(req.Amount) {
			return errors.New("INSUFFICIENT_FUNDS")
		}

		// Tính toán tiền phế và tiền thực nhận
		commissionAmount := req.Amount.Mul(commissionRate) // 15% phế
		expertAmount := req.Amount.Sub(commissionAmount)   // 85% cho bác sĩ

		// Cập nhật ví Bệnh nhân
		payerBalanceBefore := payerWallet.Balance
		payerWallet.Balance = payerWallet.Balance.Sub(req.Amount)
		tx.Save(&payerWallet)

		// 2. Khóa và cộng tiền cho Bác sĩ (85%)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", req.PayeeID).First(&expertWallet).Error; err != nil {
			return err
		}
		expertBalanceBefore := expertWallet.Balance
		expertWallet.Balance = expertWallet.Balance.Add(expertAmount)
		tx.Save(&expertWallet)

		// 3. Khóa và cộng tiền cho Hệ thống (15%)
		// Giả sử có một ví với UserType là 'SYSTEM'
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_type = ?", "SYSTEM").First(&systemWallet).Error; err != nil {
			return errors.New("Hệ thống chưa cấu hình Ví Admin")
		}
		systemBalanceBefore := systemWallet.Balance
		systemWallet.Balance = systemWallet.Balance.Add(commissionAmount)
		tx.Save(&systemWallet)

		// 4. Ghi 3 dòng Transaction để đối soát
		orderUUID, _ := uuid.Parse(req.RelatedOrderID)

		// Dòng 1: Bệnh nhân chi (-100%)
		tx.Create(&models.Transaction{WalletID: payerWallet.WalletID, RelatedOrderID: orderUUID, TransactionType: "PAYMENT", Amount: req.Amount.Neg(), BalanceBefore: payerBalanceBefore, BalanceAfter: payerWallet.Balance, Status: "SUCCESS"})

		// Dòng 2: Bác sĩ nhận (+85%)
		tx.Create(&models.Transaction{WalletID: expertWallet.WalletID, RelatedOrderID: orderUUID, TransactionType: "EARNING", Amount: expertAmount, BalanceBefore: expertBalanceBefore, BalanceAfter: expertWallet.Balance, Status: "SUCCESS"})

		// Dòng 3: Hệ thống thu phế (+15%)
		tx.Create(&models.Transaction{WalletID: systemWallet.WalletID, RelatedOrderID: orderUUID, TransactionType: "COMMISSION", Amount: commissionAmount, BalanceBefore: systemBalanceBefore, BalanceAfter: systemWallet.Balance, Status: "SUCCESS"})

		return nil
	})

	// Xử lý kết quả Transaction
	if err != nil {
		if err.Error() == "INSUFFICIENT_FUNDS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số dư trong ví không đủ để thanh toán!"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Giao dịch thất bại: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thanh toán thành công!"})
}

func RequestWithdrawal(c *gin.Context) {
	ownerID := c.Param("owner_id")
	var req schemas.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		// 1. Khóa ví
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", ownerID).First(&wallet).Error; err != nil {
			return err
		}

		// 2. Kiểm tra số dư
		if wallet.Balance.LessThan(req.Amount) {
			return errors.New("INSUFFICIENT_FUNDS")
		}

		// 3. Trừ tiền
		balanceBefore := wallet.Balance
		wallet.Balance = wallet.Balance.Sub(req.Amount)
		tx.Save(&wallet)

		// 4. Tạo phiếu Yêu cầu rút tiền (PENDING)
		withdrawalReq := models.WithdrawalRequest{
			WalletID: wallet.WalletID,
			Amount:   req.Amount,
			BankInfo: req.BankInfo,
			Status:   "PENDING",
		}
		if err := tx.Create(&withdrawalReq).Error; err != nil {
			return err
		}

		// 5. Ghi nhận Transaction (Dùng ID của yêu cầu rút làm RelatedOrderID)
		txn := models.Transaction{
			WalletID:        wallet.WalletID,
			RelatedOrderID:  withdrawalReq.RequestID,
			TransactionType: "WITHDRAW",
			Amount:          req.Amount.Neg(), // Trừ tiền
			BalanceBefore:   balanceBefore,
			BalanceAfter:    wallet.Balance,
			Status:          "PENDING", // Trạng thái giao dịch chờ duyệt
		}
		return tx.Create(&txn).Error
	})

	if err != nil {
		if err.Error() == "INSUFFICIENT_FUNDS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số dư không đủ để rút!"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu rút tiền thành công. Vui lòng chờ Admin xử lý!"})
}

// 6. ADMIN DUYỆT/TỪ CHỐI RÚT TIỀN
func ProcessWithdrawal(c *gin.Context) {
	requestID := c.Param("request_id")
	var req schemas.ProcessWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var withdrawalReq models.WithdrawalRequest
		// Tìm phiếu yêu cầu
		if err := tx.Where("request_id = ?", requestID).First(&withdrawalReq).Error; err != nil {
			return errors.New("Không tìm thấy yêu cầu rút tiền")
		}

		if withdrawalReq.Status != "PENDING" {
			return errors.New("Yêu cầu này đã được xử lý rồi!")
		}

		if req.Action == "APPROVE" {
			// NẾU DUYỆT: Chỉ cần đổi trạng thái (Tiền đã trừ từ trước rồi)
			withdrawalReq.Status = "APPROVED"
			withdrawalReq.AdminNote = req.AdminNote
			tx.Save(&withdrawalReq)

			// Cập nhật lại Transaction tương ứng thành SUCCESS
			tx.Model(&models.Transaction{}).Where("related_order_id = ?", withdrawalReq.RequestID).Update("status", "SUCCESS")

		} else if req.Action == "REJECT" {
			// NẾU TỪ CHỐI: Đổi trạng thái và HOÀN TIỀN
			withdrawalReq.Status = "REJECTED"
			withdrawalReq.AdminNote = req.AdminNote
			tx.Save(&withdrawalReq)

			// Khóa ví và hoàn tiền
			var wallet models.Wallet
			tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_id = ?", withdrawalReq.WalletID).First(&wallet)

			balanceBefore := wallet.Balance
			wallet.Balance = wallet.Balance.Add(withdrawalReq.Amount)
			tx.Save(&wallet)

			// Ghi thêm 1 dòng Transaction mới: HOÀN TIỀN
			refundTxn := models.Transaction{
				WalletID:        wallet.WalletID,
				RelatedOrderID:  withdrawalReq.RequestID,
				TransactionType: "REFUND_WITHDRAW",
				Amount:          withdrawalReq.Amount,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    wallet.Balance,
				Status:          "SUCCESS",
			}
			tx.Create(&refundTxn)
		} else {
			return errors.New("Action không hợp lệ")
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xử lý yêu cầu rút tiền thành công!"})
}
