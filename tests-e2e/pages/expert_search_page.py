from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class ExpertSearchPage(BasePage):
    SEARCH_INPUT = (By.CSS_SELECTOR, "input[aria-label='Search experts'], input[placeholder*='Search by expert']")
    SPECIALIZATION_SELECT = (By.CSS_SELECTOR, "select[aria-label='Filter by specialization']")
    EXPERT_CARDS = (By.XPATH, "//div[contains(@class, 'grid')]//h3")
    NO_RESULTS_MSG = (By.XPATH, "//p[contains(., 'No experts found') or contains(., 'không tìm thấy')]")
    FIRST_VIEW_PROFILE_BTN = (By.XPATH, "(//a[contains(., 'View profile')])[1]")

    def navigate(self):
        self.driver.get(f"{BASE_URL}/patient/find-experts")

    def search_by_name(self, name: str):
        self.type_text(self.SEARCH_INPUT, name)

    def select_specialization(self, spec_name: str):
        from selenium.webdriver.support.ui import Select
        select_element = Select(self.find(self.SPECIALIZATION_SELECT))
        select_element.select_by_visible_text(spec_name)

    def get_expert_names(self):
        elements = self.find_all(self.EXPERT_CARDS, timeout=5)
        return [el.text for el in elements if el.text]

    def is_no_results_displayed(self):
        return self.is_visible(self.NO_RESULTS_MSG, timeout=5)

    def click_first_expert_profile(self):
        self.click(self.FIRST_VIEW_PROFILE_BTN)
