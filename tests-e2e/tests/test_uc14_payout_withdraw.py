import pytest
from pages.login_page import LoginPage
from pages.wallet_page import WalletPage

@pytest.mark.wallet
def test_TC_PAY_WTD_01_open_withdraw_dialog_and_fill_bank_info(driver):
    """TC-PAY-WTD-01: Mở dialog rút tiền và nhập thông tin tài khoản ngân hàng hợp lệ (Happy Case)."""
    login_page = LoginPage(driver)
    wallet_page = WalletPage(driver)

    login_page.login_as_patient()
    wallet_page.navigate_patient_wallet()

    if wallet_page.is_setup_wallet_visible():
        wallet_page.click_setup_wallet()

    # Open withdraw dialog
    wallet_page.open_withdraw_dialog()

    # Fill valid amount and bank details
    wallet_page.fill_withdraw_form("2000000", "MB Bank - STK: 999999999 - Chu TK: NGUYEN VAN A")
    wallet_page.submit_withdraw()

    # Verify response: either success toast or dialog processed
    assert wallet_page.is_success_toast_visible(timeout=5) or not wallet_page.is_withdraw_error_visible()

@pytest.mark.wallet
def test_TC_PAY_WTD_02_validation_error_invalid_withdraw(driver):
    """TC-PAY-WTD-02: Rút tiền thất bại do thiếu thông tin ngân hàng hoặc số tiền <= 0 (Validation Error)."""
    login_page = LoginPage(driver)
    wallet_page = WalletPage(driver)

    login_page.login_as_patient()
    wallet_page.navigate_patient_wallet()

    if wallet_page.is_setup_wallet_visible():
        wallet_page.click_setup_wallet()

    wallet_page.open_withdraw_dialog()

    # Submit without bank info or with 0 amount
    wallet_page.fill_withdraw_form("0", "")
    wallet_page.submit_withdraw()

    # Verify validation error appears
    assert wallet_page.is_withdraw_error_visible(), "Validation error should appear for invalid amount / empty bank info"

@pytest.mark.wallet
def test_TC_PAY_WTD_03_admin_view_withdrawals_management(driver):
    """TC-PAY-WTD-03: Admin truy cập trang quản lý giải ngân rút tiền (/admin/withdrawals)."""
    login_page = LoginPage(driver)
    wallet_page = WalletPage(driver)

    login_page.login_as_admin()
    wallet_page.navigate_admin_withdrawals()

    # Verify admin withdrawals page loads
    assert "/admin/withdrawals" in driver.current_url
    assert wallet_page.is_admin_withdrawals_page_visible(), "Admin withdrawals management page should be displayed"
