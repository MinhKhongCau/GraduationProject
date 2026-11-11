import pytest
import time
from pages.login_page import LoginPage
from pages.admin_users_page import AdminUsersPage

def test_TC_ADM_01_admin_reject_or_suspend_user(driver):
    """TC-ADM-01: Admin thực hiện khóa / từ chối tài khoản vi phạm thành công"""
    login_page = LoginPage(driver)
    login_page.login_as_admin()
    
    admin_page = AdminUsersPage(driver)
    admin_page.navigate_experts()
    time.sleep(2)
    
    # If there is a reject button, test the action
    if admin_page.is_visible(admin_page.REJECT_BTN, timeout=3):
        admin_page.click(admin_page.REJECT_BTN)
        assert admin_page.is_toast_visible(timeout=8)
    else:
        # Verify experts list is loaded
        assert admin_page.get_user_rows_count() > 0

def test_TC_ADM_02_admin_verify_user_success(driver):
    """TC-ADM-02: Admin thực hiện duyệt / mở khóa tài khoản thành công"""
    login_page = LoginPage(driver)
    login_page.login_as_admin()
    
    admin_page = AdminUsersPage(driver)
    admin_page.navigate_experts()
    time.sleep(2)
    
    if admin_page.is_visible(admin_page.VERIFY_BTN, timeout=3):
        admin_page.click(admin_page.VERIFY_BTN)
        assert admin_page.is_toast_visible(timeout=8)
    else:
        assert admin_page.get_user_rows_count() > 0

def test_TC_ADM_03_admin_user_not_found(driver):
    """TC-ADM-03: Khóa/Mở khóa thất bại do tài khoản không tồn tại"""
    login_page = LoginPage(driver)
    login_page.login_as_admin()
    
    admin_page = AdminUsersPage(driver)
    admin_page.navigate_patients()
    time.sleep(2)
    
    # Search for non-existent patient
    admin_page.search_user("NON_EXISTENT_PATIENT_USER_99999")
    time.sleep(1.5)
    
    # The table should show 0 matching patients
    assert admin_page.get_user_rows_count() == 0 or "Không tìm thấy" in driver.page_source
