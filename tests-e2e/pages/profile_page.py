from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class ProfilePage(BasePage):
    FULL_NAME_INPUT = (By.CSS_SELECTOR, "input#profile-full-name, input[name='fullName']")
    DOB_INPUT = (By.CSS_SELECTOR, "input#profile-dob, input[name='dateOfBirth']")
    SAVE_BTN = (By.CSS_SELECTOR, "button[type='submit']")
    ERROR_MSG = (By.CSS_SELECTOR, "p.text-danger, .error-message, [role='alert'], .text-red-500")
    SUCCESS_ALERT = (By.CSS_SELECTOR, "div[role='status'], .text-success")

    def navigate_patient(self):
        self.driver.get(f"{BASE_URL}/patient/settings")

    def navigate_expert(self):
        self.driver.get(f"{BASE_URL}/expert/settings")

    def get_full_name(self):
        return self.get_value(self.FULL_NAME_INPUT)

    def get_dob(self):
        return self.get_value(self.DOB_INPUT)

    def enter_full_name(self, name: str):
        self.type_text(self.FULL_NAME_INPUT, name)

    def enter_dob(self, dob: str):
        self.type_text(self.DOB_INPUT, dob)

    def click_save(self):
        self.click(self.SAVE_BTN)

    def get_error_message(self):
        return self.get_text(self.ERROR_MSG)

    def is_success_alert_present(self, timeout=5):
        return self.is_visible(self.SUCCESS_ALERT, timeout=timeout)
