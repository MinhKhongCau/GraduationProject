from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class PaymentPage(BasePage):
    # Payment Result page locators
    SUCCESS_ICON = (By.CSS_SELECTOR, ".text-success, svg.text-success")
    RESULT_TITLE = (By.XPATH, "//h1[contains(., 'Payment') or contains(., 'Thanh toán')] | //h2")
    RETURN_HOME_BTN = (By.XPATH, "//a[contains(., 'Return') or contains(., 'Home') or contains(., 'Dashboard')]")
    
    # Booking payment button
    PAY_NOW_BTN = (By.XPATH, "//button[contains(., 'Thanh toán ngay') or contains(., 'Pay now')]")
    
    def navigate_payment_result(self, order_id: str = "TEST_ORDER"):
        self.driver.get(f"{BASE_URL}/payment-result?orderId={order_id}&status=00")

    def is_payment_success_displayed(self):
        return self.is_visible(self.SUCCESS_ICON, timeout=5) or self.is_visible(self.RESULT_TITLE, timeout=5)

    def is_vnpay_redirect(self):
        return "vnpay" in self.driver.current_url.lower() or "sandbox.vnpayment.vn" in self.driver.current_url.lower()
