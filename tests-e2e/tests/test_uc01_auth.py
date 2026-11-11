import pytest
from pages.login_page import LoginPage
from conftest import BASE_URL
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

def test_TC_AUTH_LGN_CON_01_login_success(driver):
    """TC-AUTH-LGN-CON-01: Đăng nhập thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    login_page.enter_email("patient@mindcare.com")
    login_page.enter_password("patient@mindcare.com")
    login_page.click_submit()
    
    login_page.wait_for_url_contains("/patient", timeout=10)
    assert "/patient/dashboard" in driver.current_url or "/patient" in driver.current_url

def test_TC_AUTH_LGN_CON_02_empty_email(driver):
    """TC-AUTH-LGN-CON-02: Lỗi validate: Trống email"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    login_page.enter_password("patient@mindcare.com")
    login_page.click_submit()
    
    error_msg = login_page.get_error_message()
    assert "Enter a valid email address" in error_msg

def test_TC_AUTH_LGN_CON_03_invalid_email_format(driver):
    """TC-AUTH-LGN-CON-03: Lỗi validate: Email sai định dạng"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    # 'invalid@domain' satisfies browser HTML5 email input but fails Zod EMAIL_REGEX (/^[^\s@]+@[^\s@]+\.[^\s@]+$/)
    login_page.enter_email("invalid@domain")
    login_page.enter_password("password123")
    login_page.click_submit()
    
    error_msg = login_page.get_error_message()
    assert "Enter a valid email address" in error_msg

def test_TC_AUTH_LGN_SVC_01_verify_session_tokens(driver):
    """TC-AUTH-LGN-SVC-01: Xác thực thành công (Service) - Token & Session"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    login_page.enter_email("patient@mindcare.com")
    login_page.enter_password("patient@mindcare.com")
    login_page.click_submit()
    
    login_page.wait_for_url_contains("/patient", timeout=10)
    token = login_page.get_token_from_local_storage()
    role = login_page.get_role_from_local_storage()
    
    assert token is not None and len(token) > 10
    assert role == "PATIENT"

def test_TC_AUTH_LGN_SVC_02_account_not_found(driver):
    """TC-AUTH-LGN-SVC-02: Lỗi tài khoản không tồn tại"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    login_page.enter_email("nonexistent_user_999@mindcare.com")
    login_page.enter_password("Password123!")
    login_page.click_submit()
    
    # An error notification or alert should appear, and user remains on login page
    WebDriverWait(driver, 5).until(lambda d: "/auth/login" in d.current_url)
    assert "/auth/login" in driver.current_url

def test_TC_AUTH_LGN_SVC_03_wrong_password(driver):
    """TC-AUTH-LGN-SVC-03: Lỗi sai mật khẩu"""
    login_page = LoginPage(driver)
    login_page.navigate()
    
    login_page.enter_email("patient@mindcare.com")
    login_page.enter_password("WrongPassword123!")
    login_page.click_submit()
    
    # User remains on login page
    WebDriverWait(driver, 5).until(lambda d: "/auth/login" in d.current_url)
    assert "/auth/login" in driver.current_url

def test_TC_AUTH_LGO_01_logout_success(driver):
    """TC-AUTH-LGO-01: Đăng xuất thành công (Logout)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    login_page.logout()
    login_page.wait_for_url_contains("/auth/login", timeout=10)
    assert "/auth/login" in driver.current_url
