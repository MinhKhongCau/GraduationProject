package payment

import "encoding/json"

func BookingOutboxPayload(appointmentID, orderID, status string) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"appointment_id": appointmentID,
		"order_id":       orderID,
		"status":         status,
	})
	return string(payload), err
}
