package payment

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAdminReviewAction  = errors.New("admin review action must be MANUAL_REVIEW or REFUND_REQUIRED")
	ErrOrderNotReviewable        = errors.New("only successful appointment payment orders can be sent to review")
	ErrCompensationAlreadyClosed = errors.New("compensation case is already resolved")
	ErrResolutionNoteTooLong     = errors.New("note must not exceed 500 characters")
)

const maxResolutionNoteLength = 500

// PlanAdminReview quyết định trạng thái xử lý khi Admin quản lý chuyên gia đưa một đơn đã
// thanh toán vào diện kiểm tra thủ công hoặc yêu cầu hoàn tiền.
func PlanAdminReview(order *PaymentOrder, action CompensationStatus, note string) (FulfillmentStatus, CompensationPlan, error) {
	if order.Status != OrderStatusSuccess || order.AppointmentID == nil {
		return "", CompensationPlan{}, ErrOrderNotReviewable
	}
	note = strings.TrimSpace(note)
	if len([]rune(note)) > maxResolutionNoteLength {
		return "", CompensationPlan{}, ErrResolutionNoteTooLong
	}
	plan := CompensationPlan{Type: CompensationAdminReview, Status: action, SafeReason: note}
	switch action {
	case CompensationManualReview:
		plan.ReasonCode = CompensationReasonAdminManualReview
		if plan.SafeReason == "" {
			plan.SafeReason = "Payment order was flagged for manual review by the managing administrator."
		}
		return FulfillmentManualReview, plan, nil
	case CompensationRefundRequired:
		plan.ReasonCode = CompensationReasonAdminRefundRequest
		if plan.SafeReason == "" {
			plan.SafeReason = "Refund was requested by the managing administrator."
		}
		return FulfillmentRefundRequired, plan, nil
	default:
		return "", CompensationPlan{}, ErrInvalidAdminReviewAction
	}
}

// Resolve đóng một hồ sơ bồi hoàn sau khi Admin đã xử lý xong.
func (c *PaymentCompensationCase) Resolve(note string, atMs int64) error {
	if c.Status == CompensationResolved {
		return ErrCompensationAlreadyClosed
	}
	note = strings.TrimSpace(note)
	if len([]rune(note)) > maxResolutionNoteLength {
		return ErrResolutionNoteTooLong
	}
	c.Status = CompensationResolved
	c.ResolvedAt = &atMs
	c.UpdatedAt = atMs
	if note != "" {
		c.ResolutionNote = &note
	}
	return nil
}
