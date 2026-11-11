import pytest
import time
from pages.login_page import LoginPage
from pages.wallet_page import WalletPage

def test_TC_PAY_WLT_01_top_up_wallet_success(driver):
    """TC-PAY-WLT-01: Nạp tiền vào ví điện tử thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    wallet_page = WalletPage(driver)
    wallet_page.navigate_patient_wallet()
    time.sleep(2)
    
    if wallet_page.is_setup_wallet_visible():
        wallet_page.click_setup_wallet()
        time.sleep(2)
        
    if wallet_page.is_visible(wallet_page.TOP_UP_BTN, timeout=3):
        wallet_page.top_up("50000")
        # Verify success toast notification
        assert wallet_page.is_success_toast_visible(timeout=8)

def test_TC_PAY_WLT_02_expert_earnings_breakdown(driver):
    """TC-PAY-WLT-02: Ghi nhận thù lao vào ví chuyên gia thành công"""
    login_page = LoginPage(driver)
    login_page.login_as_expert()
    
    wallet_page = WalletPage(driver)
    wallet_page.navigate_expert_wallet()
    time.sleep(2)
    
    # Verify earnings summary cards: Gross revenue, Commission (15%), Net amount (85%)
    assert wallet_page.is_expert_stats_visible()

def test_TC_PAY_WLT_03_view_wallet_balance_success(driver):
    """TC-PAY-WLT-03: Xem thông tin và số dư ví điện tử thành công"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    wallet_page = WalletPage(driver)
    wallet_page.navigate_patient_wallet()
    time.sleep(2)
    
    if wallet_page.is_setup_wallet_visible():
        wallet_page.click_setup_wallet()
        time.sleep(2)
        
    balance_text = wallet_page.get_patient_balance()
    assert "đ" in balance_text or len(balance_text) > 0
