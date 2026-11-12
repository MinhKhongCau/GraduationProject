import pytest
from pages.login_page import LoginPage
from pages.notification_page import NotificationPage

@pytest.mark.notification
def test_TC_NOTIF_01_booking_notification_received(driver):
    """TC-NOTIF-01: Gửi và hiển thị thông báo sự kiện đặt lịch hẹn thành công (Happy Case).
    
    Theo tài liệu thiết kế frontend (DESIGN.md), chức năng Notification Center 
    chưa được hoàn thiện ở phiên bản này, kịch bản kiểm thử sẽ FAIL nếu chưa hoàn thành.
    """
    login_page = LoginPage(driver)
    notif_page = NotificationPage(driver)

    login_page.login_as_patient()
    
    # Check if notification bell or notification center exists in header/navbar
    bell_present = notif_page.is_notification_bell_present(timeout=3)
    assert bell_present, (
        "TC-NOTIF-01 FAILED: Thành phần chuông thông báo (Notification Bell) "
        "chưa được tích hợp trên giao diện thanh điều hướng người dùng."
    )

@pytest.mark.notification
def test_TC_NOTIF_02_payment_notification_received(driver):
    """TC-NOTIF-02: Gửi và ghi nhận thông báo sự kiện thanh toán thành công (Happy Case).
    
    Kiểm tra danh sách thông báo thanh toán trong trung tâm thông báo hoặc trang /notifications.
    """
    login_page = LoginPage(driver)
    notif_page = NotificationPage(driver)

    login_page.login_as_patient()
    notif_page.navigate_notifications_page()

    # Check that notifications page exists and is not a 404 / redirect
    assert "/notifications" in driver.current_url, (
        "TC-NOTIF-02 FAILED: Trang thông báo hệ thống /notifications chưa được định tuyến."
    )
    items_count = notif_page.get_notification_items_count()
    assert items_count > 0, "TC-NOTIF-02 FAILED: Không tìm thấy thông báo thanh toán nào trong danh sách."

@pytest.mark.notification
def test_TC_NOTIF_03_mark_notification_as_read(driver):
    """TC-NOTIF-03: Đánh dấu thông báo đã đọc thành công."""
    login_page = LoginPage(driver)
    notif_page = NotificationPage(driver)

    login_page.login_as_patient()
    bell_present = notif_page.is_notification_bell_present(timeout=3)
    assert bell_present, (
        "TC-NOTIF-03 FAILED: Không tìm thấy giao diện thông báo để thực hiện đánh dấu đã đọc."
    )
    notif_page.open_notifications_dropdown()
    notif_page.mark_notifications_read()
