package domain

import "testing"

func TestPlanBookingDelivery(t *testing.T) {
	tests := []struct {
		name        string
		eventType   string
		outbox      OutboxStatus
		success     bool
		failed      bool
		category    BookingDeliveryFailureCategory
		fulfillment FulfillmentStatus
		caseStatus  CompensationStatus
		reason      CompensationReasonCode
	}{
		{name: "successful confirmation", eventType: BookingConfirmEvent, outbox: OutboxStatusDelivered, success: true, fulfillment: FulfillmentBookingConfirmed},
		{name: "failed payment cancellation", eventType: BookingFailEvent, outbox: OutboxStatusDelivered, failed: true, fulfillment: FulfillmentBookingFailed},
		{name: "opposite terminal conflict", eventType: BookingConfirmEvent, outbox: OutboxStatusDead, success: true, category: BookingDeliveryConflict, fulfillment: FulfillmentRefundRequired, caseStatus: CompensationRefundRequired, reason: CompensationReasonBookingConflict},
		{name: "appointment not found", eventType: BookingConfirmEvent, outbox: OutboxStatusDead, success: true, category: BookingDeliveryNotFound, fulfillment: FulfillmentRefundRequired, caseStatus: CompensationRefundRequired, reason: CompensationReasonAppointmentNotFound},
		{name: "authentication rejection", eventType: BookingConfirmEvent, outbox: OutboxStatusDead, success: true, category: BookingDeliveryAuthentication, fulfillment: FulfillmentManualReview, caseStatus: CompensationManualReview, reason: CompensationReasonBookingAuthentication},
		{name: "retries exhausted", eventType: BookingConfirmEvent, outbox: OutboxStatusDead, success: true, category: BookingDeliveryUpstream, fulfillment: FulfillmentManualReview, caseStatus: CompensationManualReview, reason: CompensationReasonBookingDeliveryRetries},
		{name: "ambiguous contract", eventType: BookingConfirmEvent, outbox: OutboxStatusDead, success: true, category: BookingDeliveryMalformedResponse, fulfillment: FulfillmentManualReview, caseStatus: CompensationManualReview, reason: CompensationReasonBookingContract},
		{name: "unrelated event", eventType: "wallet.payment.received", outbox: OutboxStatusDead, success: true},
		{name: "confirm for failed payment", eventType: BookingConfirmEvent, outbox: OutboxStatusDelivered, failed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := PlanBookingDelivery(test.eventType, test.outbox, test.success, test.failed, test.category)
			if decision.FulfillmentStatus != test.fulfillment {
				t.Fatalf("expected fulfillment %q, got %q", test.fulfillment, decision.FulfillmentStatus)
			}
			if test.caseStatus == "" {
				if decision.Compensation != nil {
					t.Fatalf("unexpected compensation: %+v", decision.Compensation)
				}
				return
			}
			if decision.Compensation == nil || decision.Compensation.Status != test.caseStatus || decision.Compensation.ReasonCode != test.reason {
				t.Fatalf("unexpected compensation: %+v", decision.Compensation)
			}
		})
	}
}

func TestPlanDuplicateGatewayCapture(t *testing.T) {
	plan := PlanDuplicateGatewayCapture()
	if plan.Type != CompensationDuplicateCapture || plan.Status != CompensationRefundRequired || plan.ReasonCode != CompensationReasonDuplicateCapture {
		t.Fatalf("unexpected duplicate capture plan: %+v", plan)
	}
}
