from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class WalletPage(BasePage):
    # Patient Wallet Locators
    BALANCE_TEXT = (By.XPATH, "//p[contains(text(), 'Total available balance')]/following-sibling::p[1]")
    TOP_UP_BTN = (By.XPATH, "//button[contains(., 'Top up')]")
    WITHDRAW_BTN = (By.XPATH, "//button[contains(., 'Withdraw')]")
    SETUP_WALLET_BTN = (By.XPATH, "//button[contains(., 'Set up my wallet')]")
    
    # Top up dialog locators
    AMOUNT_INPUT = (By.CSS_SELECTOR, "input#topup-amount, input[name='amount']")
    CONFIRM_TOP_UP_BTN = (By.XPATH, "//button[contains(., 'Confirm top up') or contains(., 'Processing...')]")
    MODAL_CONTAINER = (By.CSS_SELECTOR, "div[role='dialog']")
    
    # Withdraw dialog locators
    WITHDRAW_AMOUNT_INPUT = (By.CSS_SELECTOR, "input#withdraw-amount, input[name='amount']")
    WITHDRAW_BANK_INPUT = (By.CSS_SELECTOR, "input#withdraw-bank, input[name='bankInfo']")
    WITHDRAW_SUBMIT_BTN = (By.XPATH, "//div[@role='dialog']//button[@type='submit']")
    WITHDRAW_ERROR_MSG = (By.CSS_SELECTOR, "div[role='dialog'] p.text-danger, div[role='dialog'] [role='alert']")

    # Expert Wallet Locators
    GROSS_REVENUE_CARD = (By.XPATH, "//div[contains(., 'Gross revenue')]")
    COMMISSION_CARD = (By.XPATH, "//div[contains(., 'Commission')]")
    NET_AMOUNT_CARD = (By.XPATH, "//div[contains(., 'Net amount')]")
    ORDERS_CARD = (By.XPATH, "//div[contains(., 'Orders')]")
    EARNINGS_ROWS = (By.XPATH, "//table//tbody//tr")

    # Admin Withdrawals Locators
    ADMIN_WITHDRAWALS_TITLE = (By.XPATH, "//h1[contains(., 'Withdrawals')] | //div[contains(., 'Coming Soon')]")

    # Toast status message
    TOAST_SUCCESS = (By.CSS_SELECTOR, "div[role='status'], .text-success")

    def navigate_patient_wallet(self):
        self.driver.get(f"{BASE_URL}/patient/wallet")

    def navigate_expert_wallet(self):
        self.driver.get(f"{BASE_URL}/expert/wallet")

    def navigate_admin_withdrawals(self):
        self.driver.get(f"{BASE_URL}/admin/withdrawals")

    def get_patient_balance(self):
        return self.get_text(self.BALANCE_TEXT)

    def open_top_up_dialog(self):
        self.click(self.TOP_UP_BTN)

    def enter_top_up_amount(self, amount: str):
        self.type_text(self.AMOUNT_INPUT, amount)

    def submit_top_up(self):
        self.click(self.CONFIRM_TOP_UP_BTN)

    def top_up(self, amount: str):
        self.open_top_up_dialog()
        self.enter_top_up_amount(amount)
        self.submit_top_up()

    def open_withdraw_dialog(self):
        self.click(self.WITHDRAW_BTN)

    def fill_withdraw_form(self, amount: str, bank_info: str):
        self.type_text(self.WITHDRAW_AMOUNT_INPUT, amount)
        self.type_text(self.WITHDRAW_BANK_INPUT, bank_info)

    def submit_withdraw(self):
        self.click(self.WITHDRAW_SUBMIT_BTN)

    def is_withdraw_error_visible(self, timeout=3) -> bool:
        return self.is_visible(self.WITHDRAW_ERROR_MSG, timeout=timeout)

    def is_setup_wallet_visible(self):
        return self.is_visible(self.SETUP_WALLET_BTN, timeout=3)

    def click_setup_wallet(self):
        self.click(self.SETUP_WALLET_BTN)

    def is_expert_stats_visible(self):
        return (
            self.is_visible(self.GROSS_REVENUE_CARD, timeout=5) and
            self.is_visible(self.NET_AMOUNT_CARD, timeout=5)
        )

    def is_admin_withdrawals_page_visible(self) -> bool:
        return self.is_visible(self.ADMIN_WITHDRAWALS_TITLE, timeout=5)

    def is_success_toast_visible(self, timeout=5):
        return self.is_visible(self.TOAST_SUCCESS, timeout=timeout)
