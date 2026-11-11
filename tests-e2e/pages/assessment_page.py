from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class AssessmentPage(BasePage):
    # Admin Template Page Locators
    CREATE_TEMPLATE_BTN = (By.XPATH, "//button[contains(., 'Tạo bài test mới')]")
    TEMPLATE_ROWS = (By.XPATH, "//table//tbody//tr")
    
    # Template Modal Locators
    CODE_INPUT = (By.ID, "templatemodal-code")
    TITLE_INPUT = (By.ID, "templatemodal-title")
    DESCRIPTION_INPUT = (By.ID, "templatemodal-description")
    MODAL_SUBMIT_BTN = (By.XPATH, "//div[@role='dialog']//button[@type='submit']")
    MODAL_CLOSE_BTN = (By.XPATH, "//div[@role='dialog']//button[contains(@aria-label, 'Close') or contains(., 'Đóng') or contains(., 'Hủy')]")
    
    # Patient Assessment Locators
    ASSESSMENT_CARDS = (By.XPATH, "//div[contains(@class, 'grid')]//h3")
    START_ASSESSMENT_LINK = (By.XPATH, "(//a[contains(., 'Start assessment')])[1]")
    ASSESSMENT_DETAIL_TITLE = (By.XPATH, "//h1")

    # Toast / Alert
    TOAST_MESSAGE = (By.CSS_SELECTOR, "div[role='status'], div[role='alert']")

    def navigate_admin_templates(self):
        self.driver.get(f"{BASE_URL}/admin/templates")

    def navigate_patient_assessments(self):
        self.driver.get(f"{BASE_URL}/patient/assessment")

    def open_create_modal(self):
        self.click(self.CREATE_TEMPLATE_BTN)

    def fill_template_form(self, code: str, title: str, description: str = ""):
        self.type_text(self.CODE_INPUT, code)
        self.type_text(self.TITLE_INPUT, title)
        if description:
            self.type_text(self.DESCRIPTION_INPUT, description)

    def submit_template_form(self):
        self.click(self.MODAL_SUBMIT_BTN)

    def get_template_rows_count(self):
        return len(self.find_all(self.TEMPLATE_ROWS, timeout=5))

    def get_patient_assessment_titles(self):
        elements = self.find_all(self.ASSESSMENT_CARDS, timeout=5)
        return [e.text for e in elements if e.text]

    def click_first_assessment(self):
        self.click(self.START_ASSESSMENT_LINK)

    def is_toast_visible(self, timeout=5):
        return self.is_visible(self.TOAST_MESSAGE, timeout=timeout)
