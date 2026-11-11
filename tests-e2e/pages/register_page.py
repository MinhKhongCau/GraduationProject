from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class RegisterPage(BasePage):
    ROLE_GROUP = (By.CSS_SELECTOR, "div[role='group']")
    ROLE_PATIENT_BTN = (By.XPATH, "//div[@role='group']//button[contains(., 'Patient')]")
    ROLE_EXPERT_BTN = (By.XPATH, "//div[@role='group']//button[contains(., 'Expert')]")
    
    FULL_NAME_INPUT = (By.CSS_SELECTOR, "input[name='fullName']")
    EMAIL_INPUT = (By.CSS_SELECTOR, "input[name='email']")
    DOB_INPUT = (By.CSS_SELECTOR, "input[name='dateOfBirth']")
    PASSWORD_INPUT = (By.CSS_SELECTOR, "input[name='password']")
    CONFIRM_PASSWORD_INPUT = (By.CSS_SELECTOR, "input[name='confirmPassword']")
    SUBMIT_BTN = (By.CSS_SELECTOR, "button[type='submit']")
    ERROR_MSG = (By.CSS_SELECTOR, "p.text-danger, .error-message, [role='alert'], .text-red-500")
    STATUS_MSG = (By.CSS_SELECTOR, "div[role='status'], p[role='status']")

    def navigate(self):
        self.driver.get(f"{BASE_URL}/auth/register")

    def select_role(self, role: str):
        if role.upper() == "EXPERT":
            self.click(self.ROLE_EXPERT_BTN)
        else:
            self.click(self.ROLE_PATIENT_BTN)

    def is_role_active(self, role: str) -> bool:
        btn_locator = self.ROLE_EXPERT_BTN if role.upper() == "EXPERT" else self.ROLE_PATIENT_BTN
        btn = self.find(btn_locator)
        return btn.get_attribute("aria-pressed") == "true"

    def enter_full_name(self, name: str):
        self.type_text(self.FULL_NAME_INPUT, name)

    def enter_email(self, email: str):
        self.type_text(self.EMAIL_INPUT, email)

    def enter_dob(self, dob: str):
        self.type_text(self.DOB_INPUT, dob)

    def enter_password(self, password: str):
        self.type_text(self.PASSWORD_INPUT, password)

    def enter_confirm_password(self, password: str):
        self.type_text(self.CONFIRM_PASSWORD_INPUT, password)

    def click_submit(self, bypass_html5=False):
        if bypass_html5:
            self.execute_script("const forms = document.querySelectorAll('form'); forms.forEach(f => f.noValidate = true);")
        self.click(self.SUBMIT_BTN)

    def fill_form(self, full_name="", email="", dob="", password="", confirm_password="", confirmPassword="", role="PATIENT"):
        cp = confirm_password or confirmPassword
        if role:
            self.select_role(role)
        if full_name:
            self.enter_full_name(full_name)
        if email:
            self.enter_email(email)
        if dob:
            self.enter_dob(dob)
        if password:
            self.enter_password(password)
        if cp:
            self.enter_confirm_password(cp)

    def get_error_message(self):
        return self.get_text(self.ERROR_MSG)

    def get_all_error_messages(self):
        elements = self.find_all(self.ERROR_MSG, timeout=3)
        return [e.text for e in elements if e.text]

    def get_status_message(self):
        return self.get_text(self.STATUS_MSG)
