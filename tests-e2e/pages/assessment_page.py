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
    
    # Patient Assessment List Locators
    ASSESSMENT_CARDS = (By.XPATH, "//div[contains(@class, 'grid')]//h3")
    START_ASSESSMENT_LINK = (By.XPATH, "(//a[contains(., 'Start assessment')])[1]")
    ASSESSMENT_DETAIL_TITLE = (By.XPATH, "//h1")
    MY_HISTORY_SECTION = (By.XPATH, "//h2[contains(text(), 'My Assessment History')]")

    # Patient Assessment Detail / Taking Test Locators
    START_ASSESSMENT_BTN = (By.XPATH, "//button[contains(., 'Bắt đầu đánh giá')]")
    QUESTION_CONTAINER = (By.XPATH, "//div[contains(@class, 'space-y-4')]//h2 | //div[contains(@class, 'space-y-4')]//p[contains(@class, 'font-medium')]")
    OPTION_BUTTONS = (By.CSS_SELECTOR, "div.space-y-2 button[aria-pressed]")
    NEXT_BTN = (By.XPATH, "//button[contains(., 'Tiếp theo')]")
    BACK_BTN = (By.XPATH, "//button[contains(., 'Quay lại')]")
    SUBMIT_ASSESSMENT_BTN = (By.XPATH, "//button[contains(., 'Hoàn thành')]")
    NOT_FOUND_MSG = (By.XPATH, "//*[contains(text(), 'Không tìm thấy thông tin bài đánh giá') or contains(text(), 'not found')]")

    # Result Summary Locators
    RESULT_TOTAL_SCORE = (By.XPATH, "//*[contains(text(), 'Total score')]")
    RESULT_DIMENSION_SCORES = (By.XPATH, "//*[contains(text(), 'Scores by dimension')]")
    RESULT_AI_EVALUATION = (By.XPATH, "//*[contains(text(), 'AI evaluation')]")
    BACK_TO_ASSESSMENTS_BTN = (By.XPATH, "//a[contains(., 'Back to assessments')]")

    # Toast / Alert
    TOAST_MESSAGE = (By.CSS_SELECTOR, "div[role='status'], div[role='alert']")

    def navigate_admin_templates(self):
        self.driver.get(f"{BASE_URL}/admin/templates")

    def navigate_patient_assessments(self):
        self.driver.get(f"{BASE_URL}/patient/assessment")

    def navigate_patient_assessment_detail(self, template_slug: str):
        self.driver.get(f"{BASE_URL}/patient/assessment/{template_slug}")

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

    def is_start_button_visible(self, timeout=5) -> bool:
        return self.is_visible(self.START_ASSESSMENT_BTN, timeout=timeout)

    def click_start_assessment(self):
        self.click(self.START_ASSESSMENT_BTN)

    def select_option(self, index: int = 0):
        self.wait_until_visible(self.OPTION_BUTTONS, timeout=8)
        options = self.find_all(self.OPTION_BUTTONS, timeout=5)
        if len(options) > index:
            self.scroll_into_view(options[index])
            self.driver.execute_script("arguments[0].click();", options[index])
            try:
                self.wait_until_visible((By.CSS_SELECTOR, "button[aria-pressed='true']"), timeout=5)
            except Exception:
                pass

    def is_next_btn_disabled(self) -> bool:
        btn = self.find(self.NEXT_BTN, timeout=3)
        if not btn:
            btn = self.find(self.SUBMIT_ASSESSMENT_BTN, timeout=3)
        return (btn.get_attribute("disabled") is not None) if btn else True

    def click_next_question(self):
        self.click(self.NEXT_BTN)

    def click_complete_assessment(self):
        self.click(self.SUBMIT_ASSESSMENT_BTN)

    def is_result_summary_visible(self, timeout=8) -> bool:
        return self.is_visible(self.RESULT_TOTAL_SCORE, timeout=timeout)

    def is_not_found_visible(self, timeout=10) -> bool:
        if self.is_visible(self.NOT_FOUND_MSG, timeout=timeout):
            return True
        try:
            body_text = self.driver.find_element(By.TAG_NAME, "body").text
            return "Không tìm thấy" in body_text or "not found" in body_text.lower()
        except Exception:
            return False

    def is_history_section_visible(self, timeout=5) -> bool:
        return self.is_visible(self.MY_HISTORY_SECTION, timeout=timeout)

    def is_toast_visible(self, timeout=5):
        return self.is_visible(self.TOAST_MESSAGE, timeout=timeout)
