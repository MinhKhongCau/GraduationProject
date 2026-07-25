package sysadmin_test

import (
	"testing"

	"forum-service/internal/sysadmin" // Or internal/sysadmin
)

func TestNotificationService(t *testing.T) {
	t.Run("TC-NOTIF-01 - Gửi thông báo sự kiện đặt lịch hẹn thành công (Happy Case)", func(t *testing.T) {
		svc := sysadmin.NewNotificationService(true)
		err := svc.PublishAppointmentNotification("patient-uuid-1", 100)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("TC-NOTIF-02 - Gửi thông báo sự kiện thanh toán thành công (Happy Case)", func(t *testing.T) {
		svc := sysadmin.NewNotificationService(true)
		err := svc.PublishPaymentNotification("patient-uuid-1", "ORD-888", 500000)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("TC-NOTIF-03 - Đánh dấu thông báo đã đọc thành công", func(t *testing.T) {
		svc := sysadmin.NewNotificationService(true)
		_ = svc.PublishAppointmentNotification("patient-uuid-1", 5)

		err := svc.MarkAsRead(5, "patient-uuid-1")
		if err != nil {
			t.Fatalf("expected mark as read success, got %v", err)
		}
	})

	t.Run("TC-NOTIF-04 - Xử lý thất bại khi kết nối hàng đợi tin nhắn bị ngắt", func(t *testing.T) {
		svc := sysadmin.NewNotificationService(false) // Disconnected MQ
		err := svc.PublishAppointmentNotification("patient-uuid-1", 100)
		if err != sysadmin.ErrBrokerUnavailable {
			t.Fatalf("expected ErrBrokerUnavailable, got %v", err)
		}
	})

	t.Run("TC-NOTIF-05 - Từ chối gửi thông báo do payload sự kiện bị rỗng hoặc sai định dạng", func(t *testing.T) {
		svc := sysadmin.NewNotificationService(true)
		err := svc.PublishAppointmentNotification("", 0) // Invalid payload
		if err != sysadmin.ErrInvalidNotificationPayload {
			t.Fatalf("expected ErrInvalidNotificationPayload, got %v", err)
		}
	})
}
