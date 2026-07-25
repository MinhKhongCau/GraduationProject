package sysadmin_test

import (
	"testing"

	"forum-service/internal/sysadmin"
)

func TestServiceCatalog(t *testing.T) {
	t.Run("TC-SVC-01 - Tạo dịch vụ tư vấn mới thành công (Happy Case)", func(t *testing.T) {
		catalog := sysadmin.NewServiceCatalog()
		item, err := catalog.Create("ADMIN", "Tư vấn tâm lý trực tuyến 1:1", 300000, 60, "Gặp gỡ chuyên gia qua video call")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if item.Name != "Tư vấn tâm lý trực tuyến 1:1" || item.Price != 300000 {
			t.Fatalf("unexpected item data: %+v", item)
		}
	})

	t.Run("TC-SVC-02 - Tạo dịch vụ thất bại do tên bị trống hoặc giá âm", func(t *testing.T) {
		catalog := sysadmin.NewServiceCatalog()
		_, err := catalog.Create("ADMIN", "", -50000, 60, "Mô tả")
		if err != sysadmin.ErrInvalidServiceInput {
			t.Fatalf("expected ErrInvalidServiceInput, got %v", err)
		}
	})

	t.Run("TC-SVC-03 - Cập nhật giá và mô tả dịch vụ tư vấn thành công", func(t *testing.T) {
		catalog := sysadmin.NewServiceCatalog()
		item, _ := catalog.Create("ADMIN", "Dịch vụ cũ", 300000, 60, "Mô tả cũ")

		updated, err := catalog.Update("ADMIN", item.ID, 350000, "Mô tả mới")
		if err != nil {
			t.Fatalf("expected update success, got %v", err)
		}
		if updated.Price != 350000 {
			t.Fatalf("expected updated price 350000, got %f", updated.Price)
		}
	})

	t.Run("TC-SVC-04 - Vô hiệu hóa (Soft Delete) dịch vụ tư vấn thành công", func(t *testing.T) {
		catalog := sysadmin.NewServiceCatalog()
		item, _ := catalog.Create("ADMIN", "Dịch vụ sắp xóa", 200000, 30, "Mô tả")

		err := catalog.Disable("ADMIN", item.ID)
		if err != nil {
			t.Fatalf("expected disable success, got %v", err)
		}
	})

	t.Run("TC-SVC-05 - Cập nhật dịch vụ thất bại do dịch vụ không tồn tại", func(t *testing.T) {
		catalog := sysadmin.NewServiceCatalog()
		_, err := catalog.Update("ADMIN", 99999, 500000, "Mô tả")
		if err != sysadmin.ErrServiceNotFound {
			t.Fatalf("expected ErrServiceNotFound, got %v", err)
		}
	})
}
