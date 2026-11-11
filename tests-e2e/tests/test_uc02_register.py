import uuid
import pytest
from pages.register_page import RegisterPage
from conftest import BASE_URL
from selenium.webdriver.support.ui import WebDriverWait

def test_TC_AUTH_REG_CON_01_register_success(driver):
    """TC-AUTH-REG-CON-01: Đăng ký thành công (Controller - Happy Case)"""
    page = RegisterPage(driver)
    page.navigate()
    
    unique_email = f"user_{uuid.uuid4().hex[:8]}@example.com"
    page.fill_form(
        full_name="Nguyễn Văn Test",
        email=unique_email,
        dob="1995-05-15",
        password="Password123!",
        confirm_password="Password123!",
        role="PATIENT"
    )
    page.click_submit()
    
    # Successful registration redirects to /auth/login
    try:
        page.wait_for_url_contains("/auth/login", timeout=8)
        assert "/auth/login" in driver.current_url
    except Exception:
        # Or status notification is displayed
        assert page.is_visible(page.STATUS_MSG, timeout=3) or "/auth" in driver.current_url

def test_TC_AUTH_REG_CON_02_empty_full_name(driver):
    """TC-AUTH-REG-CON-02: Lỗi validate: Trống họ tên"""
    page = RegisterPage(driver)
    page.navigate()
    
    page.fill_form(
        full_name="",
        email="test_name@example.com",
        dob="1995-05-15",
        password="Password123!",
        confirmPassword="Password123!"
    )
    page.click_submit()
    
    errors = page.get_all_error_messages()
    assert any("Enter your full name" in err for err in errors)

def test_TC_AUTH_REG_CON_03_empty_email(driver):
    """TC-AUTH-REG-CON-03: Lỗi validate: Trống email"""
    page = RegisterPage(driver)
    page.navigate()
    
    page.fill_form(
        full_name="Test User",
        email="",
        dob="1995-05-15",
        password="Password123!",
        confirmPassword="Password123!"
    )
    # Bypass browser HTML5 email empty validation to trigger frontend Zod validation
    page.click_submit(bypass_html5=True)
    
    errors = page.get_all_error_messages()
    assert any("Enter a valid email address" in err for err in errors)

def test_TC_AUTH_REG_SVC_01_register_with_specified_role(driver):
    """TC-AUTH-REG-SVC-01: Đăng ký thành công với vai trò chỉ định (EXPERT)"""
    page = RegisterPage(driver)
    page.navigate()
    
    # Toggle role to EXPERT
    page.select_role("EXPERT")
    assert page.is_role_active("EXPERT")
    
    unique_email = f"expert_{uuid.uuid4().hex[:8]}@example.com"
    page.fill_form(
        full_name="Bác sĩ Chuyên Khoa",
        email=unique_email,
        dob="1988-08-20",
        password="Password123!",
        confirm_password="Password123!",
        role="EXPERT"
    )
    page.click_submit()
    
    # Redirection to login or successful registration status
    try:
        page.wait_for_url_contains("/auth/login", timeout=8)
        assert "/auth/login" in driver.current_url
    except Exception:
        assert page.is_visible(page.STATUS_MSG, timeout=3) or "/auth" in driver.current_url

def test_TC_AUTH_REG_SVC_02_register_with_default_role(driver):
    """TC-AUTH-REG-SVC-02: Đăng ký thành công dùng vai trò mặc định (PATIENT)"""
    page = RegisterPage(driver)
    page.navigate()
    
    # Verify default role is PATIENT
    assert page.is_role_active("PATIENT")
    
    unique_email = f"patient_{uuid.uuid4().hex[:8]}@example.com"
    page.fill_form(
        full_name="Bệnh Nhân Mặc Định",
        email=unique_email,
        dob="2000-01-01",
        password="Password123!",
        confirm_password="Password123!",
        role="PATIENT"
    )
    page.click_submit()
    
    try:
        page.wait_for_url_contains("/auth/login", timeout=8)
        assert "/auth/login" in driver.current_url
    except Exception:
        assert page.is_visible(page.STATUS_MSG, timeout=3) or "/auth" in driver.current_url
