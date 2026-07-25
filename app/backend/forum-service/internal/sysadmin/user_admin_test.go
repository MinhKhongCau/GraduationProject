package sysadmin_test

import (
	"testing"

	"forum-service/internal/sysadmin"
)

func TestUserAdminService(t *testing.T) {
	t.Run("TC-ADM-01 - Admin thực hiện khóa tài khoản vi phạm thành công", func(t *testing.T) {
		svc := sysadmin.NewUserAdminService()
		svc.SeedUser(&sysadmin.UserAccount{UserID: "user-violator-uuid", Email: "violator@example.com", Role: "PATIENT", IsBlocked: false})

		err := svc.BlockUser("ADMIN", "user-violator-uuid", "Spam nội dung vi phạm tiêu chuẩn cộng đồng")
		if err != nil {
			t.Fatalf("expected block success, got %v", err)
		}
	})

	t.Run("TC-ADM-02 - Admin thực hiện mở khóa tài khoản thành công", func(t *testing.T) {
		svc := sysadmin.NewUserAdminService()
		svc.SeedUser(&sysadmin.UserAccount{UserID: "user-blocked-uuid", Email: "blocked@example.com", Role: "PATIENT", IsBlocked: true, Reason: "Spam"})

		err := svc.UnblockUser("ADMIN", "user-blocked-uuid")
		if err != nil {
			t.Fatalf("expected unblock success, got %v", err)
		}
	})

	t.Run("TC-ADM-03 - Thao tác khóa/mở khóa thất bại do tài khoản không tồn tại", func(t *testing.T) {
		svc := sysadmin.NewUserAdminService()
		err := svc.BlockUser("ADMIN", "non-existent-uuid", "Lỗi tài khoản")
		if err != sysadmin.ErrUserNotFound {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("TC-ADM-04 - Admin nâng cấp phân quyền vai trò cho người dùng thành công", func(t *testing.T) {
		svc := sysadmin.NewUserAdminService()
		svc.SeedUser(&sysadmin.UserAccount{UserID: "user-promoted-uuid", Email: "doctor@example.com", Role: "PATIENT"})

		err := svc.UpdateRole("ADMIN", "user-promoted-uuid", "EXPERT")
		if err != nil {
			t.Fatalf("expected update role success, got %v", err)
		}
	})

	t.Run("TC-ADM-05 - Từ chối thao tác quản lý người dùng nếu không phải Admin", func(t *testing.T) {
		svc := sysadmin.NewUserAdminService()
		svc.SeedUser(&sysadmin.UserAccount{UserID: "some-user-uuid", Email: "user@example.com", Role: "PATIENT"})

		err := svc.BlockUser("PATIENT", "some-user-uuid", "Hack attempt")
		if err != sysadmin.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
