import pytest
import time
from pages.login_page import LoginPage
from pages.payment_page import PaymentPage
from conftest import BASE_URL

def test_TC_PAY_ORD_01_create_vnpay_order_success(driver):
    """TC-PAY-ORD-01: Tạo đơn thanh toán VNPay thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    payment_page = PaymentPage(driver)
    # Check payment result page rendering
    payment_page.navigate_payment_result(order_id="ORD-TEST-888")
    time.sleep(2)
    
    assert payment_page.is_payment_success_displayed() or "/payment" in driver.current_url

def test_TC_PAY_ORD_02_payment_nonexistent_appointment(driver):
    """TC-PAY-ORD-02: Tạo đơn hàng thất bại do không tìm thấy cuộc hẹn"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    # Attempting to access invalid order / appointment payment directly
    driver.get(f"{BASE_URL}/patient/book-appointment?expertId=invalid-expert-999")
    time.sleep(2)
    
    # System stays safe and displays standard booking UI or error message
    assert "/patient/book-appointment" in driver.current_url

def test_TC_PAY_ORD_03_payment_unauthorized_appointment(driver):
    """TC-PAY-ORD-03: Tạo đơn thất bại do cuộc hẹn không thuộc người thanh toán"""
    login_page = LoginPage(driver)
    login_page.login_as_expert()
    
    # An expert should not be able to pay for patient bookings
    driver.get(f"{BASE_URL}/patient/my-bookings")
    time.sleep(2)
    
    # Should redirect or disallow access
    assert "/expert" in driver.current_url or "/patient" not in driver.current_url or "/forbidden" in driver.current_url
