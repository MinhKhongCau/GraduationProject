package payment

type FulfillmentStatus string

const (
	FulfillmentPending          FulfillmentStatus = "PENDING"
	FulfillmentBookingConfirmed FulfillmentStatus = "BOOKING_CONFIRMED"
	FulfillmentBookingFailed    FulfillmentStatus = "BOOKING_FAILED"
	FulfillmentManualReview     FulfillmentStatus = "MANUAL_REVIEW"
	FulfillmentRefundRequired   FulfillmentStatus = "REFUND_REQUIRED"
)

type GatewayCaptureStatus string

const (
	GatewayCapturePending   GatewayCaptureStatus = "PENDING"
	GatewayCaptureSucceeded GatewayCaptureStatus = "CAPTURED"
	GatewayCaptureFailed    GatewayCaptureStatus = "FAILED"
	GatewayCaptureDuplicate GatewayCaptureStatus = "CAPTURED_DUPLICATE"
)

type CompensationType string

const (
	CompensationBookingFulfillment CompensationType = "BOOKING_FULFILLMENT"
	CompensationDuplicateCapture   CompensationType = "DUPLICATE_GATEWAY_CAPTURE"
	CompensationAdminReview        CompensationType = "ADMIN_REVIEW"
)

type CompensationStatus string

const (
	CompensationManualReview   CompensationStatus = "MANUAL_REVIEW"
	CompensationRefundRequired CompensationStatus = "REFUND_REQUIRED"
	CompensationResolved       CompensationStatus = "RESOLVED"
)

type CompensationReasonCode string

const (
	CompensationReasonBookingConflict        CompensationReasonCode = "BOOKING_CONFLICT"
	CompensationReasonAppointmentNotFound    CompensationReasonCode = "APPOINTMENT_NOT_FOUND"
	CompensationReasonBookingAuthentication  CompensationReasonCode = "BOOKING_AUTHENTICATION_REJECTED"
	CompensationReasonBookingDeliveryRetries CompensationReasonCode = "BOOKING_DELIVERY_RETRIES_EXHAUSTED"
	CompensationReasonBookingContract        CompensationReasonCode = "BOOKING_CONTRACT_FAILURE"
	CompensationReasonDuplicateCapture       CompensationReasonCode = "DUPLICATE_GATEWAY_CAPTURE"
	CompensationReasonAdminManualReview      CompensationReasonCode = "ADMIN_MANUAL_REVIEW"
	CompensationReasonAdminRefundRequest     CompensationReasonCode = "ADMIN_REFUND_REQUEST"
)

type BookingDeliveryFailureCategory string

const (
	BookingDeliveryNetwork           BookingDeliveryFailureCategory = "network"
	BookingDeliveryTimeout           BookingDeliveryFailureCategory = "timeout"
	BookingDeliveryRateLimited       BookingDeliveryFailureCategory = "rate_limited"
	BookingDeliveryUpstream          BookingDeliveryFailureCategory = "upstream"
	BookingDeliveryMalformedResponse BookingDeliveryFailureCategory = "malformed_response"
	BookingDeliveryNotFound          BookingDeliveryFailureCategory = "not_found"
	BookingDeliveryConflict          BookingDeliveryFailureCategory = "conflict"
	BookingDeliveryAuthentication    BookingDeliveryFailureCategory = "authentication"
	BookingDeliveryBadRequest        BookingDeliveryFailureCategory = "bad_request"
	BookingDeliveryBusinessRejection BookingDeliveryFailureCategory = "business_rejection"
)

const (
	BookingConfirmEvent = "booking.appointment.confirm"
	BookingFailEvent    = "booking.appointment.fail"
)

type CompensationPlan struct {
	Type       CompensationType
	Status     CompensationStatus
	ReasonCode CompensationReasonCode
	SafeReason string
}

type BookingDeliveryDecision struct {
	FulfillmentStatus FulfillmentStatus
	Compensation      *CompensationPlan
}

func PlanBookingDelivery(
	eventType string,
	outboxStatus OutboxStatus,
	paymentSucceeded bool,
	paymentFailed bool,
	failureCategory BookingDeliveryFailureCategory,
) BookingDeliveryDecision {
	switch {
	case eventType == BookingConfirmEvent && outboxStatus == OutboxStatusDelivered && paymentSucceeded:
		return BookingDeliveryDecision{FulfillmentStatus: FulfillmentBookingConfirmed}
	case eventType == BookingFailEvent && outboxStatus == OutboxStatusDelivered && paymentFailed:
		return BookingDeliveryDecision{FulfillmentStatus: FulfillmentBookingFailed}
	case eventType != BookingConfirmEvent || outboxStatus != OutboxStatusDead || !paymentSucceeded:
		return BookingDeliveryDecision{}
	}

	plan := CompensationPlan{Type: CompensationBookingFulfillment}
	switch failureCategory {
	case BookingDeliveryConflict:
		plan.Status = CompensationRefundRequired
		plan.ReasonCode = CompensationReasonBookingConflict
		plan.SafeReason = "Booking cannot be confirmed because it is already in an opposite terminal state."
	case BookingDeliveryNotFound:
		plan.Status = CompensationRefundRequired
		plan.ReasonCode = CompensationReasonAppointmentNotFound
		plan.SafeReason = "Booking appointment was not found after payment succeeded."
	case BookingDeliveryAuthentication:
		plan.Status = CompensationManualReview
		plan.ReasonCode = CompensationReasonBookingAuthentication
		plan.SafeReason = "Booking confirmation was permanently rejected by service authentication."
	case BookingDeliveryNetwork, BookingDeliveryTimeout, BookingDeliveryRateLimited, BookingDeliveryUpstream:
		plan.Status = CompensationManualReview
		plan.ReasonCode = CompensationReasonBookingDeliveryRetries
		plan.SafeReason = "Booking confirmation exhausted delivery retries."
	default:
		plan.Status = CompensationManualReview
		plan.ReasonCode = CompensationReasonBookingContract
		plan.SafeReason = "Booking confirmation ended with an ambiguous contract or legacy failure."
	}

	fulfillment := FulfillmentManualReview
	if plan.Status == CompensationRefundRequired {
		fulfillment = FulfillmentRefundRequired
	}
	return BookingDeliveryDecision{FulfillmentStatus: fulfillment, Compensation: &plan}
}

func PlanDuplicateGatewayCapture() CompensationPlan {
	return CompensationPlan{
		Type:       CompensationDuplicateCapture,
		Status:     CompensationRefundRequired,
		ReasonCode: CompensationReasonDuplicateCapture,
		SafeReason: "A second VNPay capture was reported for an appointment that already has a successful payment.",
	}
}
