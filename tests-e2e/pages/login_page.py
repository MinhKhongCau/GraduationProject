from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class LoginPage(BasePage):
    # Locators
    EMAIL_INPUT = (By.CSS_SELECTOR, "input[type='email'], input[name='email']")
    PASSWORD_INPUT = (By.CSS_SELECTOR, "input[type='password'], input[name='password']")
    SUBMIT_BTN = (By.CSS_SELECTOR, "button[type='submit']")
    ERROR_MSG = (By.CSS_SELECTOR, "p.text-danger, .error-message, [role='alert'], .text-red-500, p[role='status']")
    
    # Navigation / Header Locators
    NAV_AVATAR_BTN = (By.CSS_SELECTOR, "button[aria-label='User menu'], button[data-state], header button.rounded-full")
    LOGOUT_BTN = (By.XPATH, "//button[contains(text(), 'Log out') or contains(text(), 'Đăng xuất')] | //div[contains(text(), 'Log out') or contains(text(), 'Đăng xuất')]")
    ADMIN_LOGOUT_BTN = (By.XPATH, "//button[contains(., 'Log out') or contains(., 'Đăng xuất')]")

    def navigate(self):
        self.driver.get(f"{BASE_URL}/auth/login")

    def enter_email(self, email):
        self.type_text(self.EMAIL_INPUT, email)

    def enter_password(self, password):
        self.type_text(self.PASSWORD_INPUT, password)

    def click_submit(self, bypass_html5=False):
        if bypass_html5:
            self.execute_script("const forms = document.querySelectorAll('form'); forms.forEach(f => f.noValidate = true);")
        self.click(self.SUBMIT_BTN)

    def login(self, email, password):
        self.navigate()
        self.enter_email(email)
        self.enter_password(password)
        self.click_submit()

    def login_as_patient(self):
        self.navigate()
        if "/patient" in self.driver.current_url:
            return
        self.enter_email("patient@mindcare.com")
        self.enter_password("patient@mindcare.com")
        self.click_submit()
        self.wait_for_url_contains("/patient", timeout=15)

    def login_as_expert(self):
        self.navigate()
        if "/expert" in self.driver.current_url:
            return
        self.enter_email("expert@mindcare.com")
        self.enter_password("expert@mindcare.com")
        self.click_submit()
        self.wait_for_url_contains("/expert", timeout=15)

    def login_as_admin(self):
        self.navigate()
        if "/admin" in self.driver.current_url:
            return
        self.enter_email("admin@mindcare.com")
        self.enter_password("admin@mindcare.com")
        self.click_submit()
        self.wait_for_url_contains("/admin", timeout=15)

    def logout(self):
        # Admin shell has direct logout button in topbar
        if self.is_visible(self.ADMIN_LOGOUT_BTN, timeout=2):
            self.click(self.ADMIN_LOGOUT_BTN)
        elif self.is_visible(self.NAV_AVATAR_BTN, timeout=3):
            self.click(self.NAV_AVATAR_BTN)
            self.click(self.LOGOUT_BTN)
        else:
            # Fallback clear auth via javascript
            self.execute_script("window.localStorage.clear(); document.cookie.split(';').forEach(c => { document.cookie = c.replace(/^ +/, '').replace(/=.*/, '=;expires=' + new Date().toUTCString() + ';path=/'); }); window.location.href = '/auth/login';")

    def get_error_message(self):
        return self.get_text(self.ERROR_MSG)

    def get_token_from_local_storage(self):
        return self.execute_script("return window.localStorage.getItem('mc_access_token');")

    def get_role_from_local_storage(self):
        return self.execute_script("return window.localStorage.getItem('mc_role');")
