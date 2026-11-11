from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class AdminUsersPage(BasePage):
    # Patients page locators
    SEARCH_INPUT = (By.CSS_SELECTOR, "input[aria-label='Tìm theo tên'], input[placeholder*='Tìm theo tên']")
    TABLE_ROWS = (By.XPATH, "//table//tbody//tr")
    PATIENT_NAME_CELLS = (By.XPATH, "//table//tbody//tr/td[1]")
    
    # Experts page locators
    EXPERT_SEARCH_INPUT = (By.CSS_SELECTOR, "input[placeholder*='Tìm theo tên'], input[aria-label='Tìm theo tên']")
    SCOPE_SELECT = (By.CSS_SELECTOR, "select[aria-label='Phạm vi']")
    VERIFY_BTN = (By.XPATH, "//button[contains(., 'Duyệt')]")
    REJECT_BTN = (By.XPATH, "//button[contains(., 'Từ chối')]")
    EDIT_BTN = (By.XPATH, "//button[contains(., 'Sửa')]")
    STATUS_BADGES = (By.XPATH, "//table//tbody//tr//span[contains(@class, 'rounded')]")
    
    # Modal
    MODAL_DIALOG = (By.CSS_SELECTOR, "div[role='dialog']")
    MODAL_SAVE_BTN = (By.XPATH, "//div[@role='dialog']//button[@type='submit' or contains(., 'Lưu')]")
    
    # Toast notification
    TOAST_NOTIFICATION = (By.CSS_SELECTOR, "div[role='status'], div[role='alert']")

    def navigate_patients(self):
        self.driver.get(f"{BASE_URL}/admin/patients")

    def navigate_experts(self):
        self.driver.get(f"{BASE_URL}/admin/experts")

    def search_user(self, name: str):
        self.type_text(self.SEARCH_INPUT, name)

    def get_user_rows_count(self):
        rows = self.find_all(self.TABLE_ROWS, timeout=5)
        return len(rows)

    def click_first_edit(self):
        self.click(self.EDIT_BTN)

    def is_edit_modal_open(self):
        return self.is_visible(self.MODAL_DIALOG, timeout=5)

    def is_toast_visible(self, timeout=5):
        return self.is_visible(self.TOAST_NOTIFICATION, timeout=timeout)

    def get_toast_message(self):
        return self.get_text(self.TOAST_NOTIFICATION)
