from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import Select
from pages.base_page import BasePage
from conftest import BASE_URL

class ForumPage(BasePage):
    # Forum list & navigation
    NEW_POST_BTN = (By.XPATH, "//a[contains(., 'New post')]")
    POST_CARDS = (By.XPATH, "//article | //div[contains(@class, 'grid')]//div[contains(@class, 'rounded')]//h2 | //div[contains(@class, 'grid')]//a")
    SEARCH_INPUT = (By.CSS_SELECTOR, "input[placeholder*='Search posts'], input[type='search'], input[aria-label*='Search']")
    CATEGORY_FILTER = (By.CSS_SELECTOR, "select[aria-label*='category' i], div[data-testid='category-filter'] select, select")

    # Post Form locators
    CATEGORY_SELECT = (By.CSS_SELECTOR, "select#post-category, select[name='categoryId']")
    TITLE_INPUT = (By.CSS_SELECTOR, "input#post-title, input[name='title']")
    SUMMARY_INPUT = (By.CSS_SELECTOR, "input#post-summary, input[name='summary']")
    CONTENT_EDITOR = (By.CSS_SELECTOR, "div[contenteditable='true'], textarea[name='content'], [placeholder*='Share your thoughts']")
    SUBMIT_BTN = (By.XPATH, "//button[@type='submit' and (contains(., 'Publish') or contains(., 'Save') or contains(., 'Saving'))]")
    ERROR_MSG = (By.CSS_SELECTOR, "p.text-danger, .error-message, [role='alert'], span.text-destructive, p.text-sm.text-danger")

    # My posts locators
    MY_POSTS_CARDS = (By.XPATH, "//div[contains(@class, 'p-5')]")
    MY_POST_FIRST_TITLE = (By.XPATH, "(//div[contains(@class, 'p-5')]//a)[1]")
    EDIT_POST_BTN = (By.XPATH, "//a[contains(., 'Edit')]")
    BLOCK_POST_BTN = (By.XPATH, "//button[contains(., 'Block')]")
    
    # Admin forum posts
    ADMIN_STATUS_FILTER = (By.CSS_SELECTOR, "select")
    ADMIN_TABLE_ROWS = (By.XPATH, "//table//tbody//tr")

    def navigate_feed(self):
        self.driver.get(f"{BASE_URL}/patient/forum")

    def navigate_my_posts(self):
        self.driver.get(f"{BASE_URL}/patient/forum/my-posts")

    def navigate_new_post(self):
        self.driver.get(f"{BASE_URL}/patient/forum/new")

    def navigate_admin_posts(self):
        self.driver.get(f"{BASE_URL}/admin/forum-posts")

    def enter_title(self, title: str):
        self.type_text(self.TITLE_INPUT, title)

    def select_first_category(self):
        sel_elem = self.find(self.CATEGORY_SELECT)
        if sel_elem:
            sel = Select(sel_elem)
            options = sel.options
            if len(options) > 1:
                sel.select_by_index(1)

    def enter_summary(self, summary: str):
        self.type_text(self.SUMMARY_INPUT, summary)

    def enter_content(self, content: str):
        editor = self.find(self.CONTENT_EDITOR, timeout=5)
        if editor:
            self.type_text(self.CONTENT_EDITOR, content)

    def click_submit(self):
        self.click(self.SUBMIT_BTN)

    def is_validation_error_displayed(self, timeout=3) -> bool:
        return self.is_visible(self.ERROR_MSG, timeout=timeout)

    def search_posts(self, query: str):
        self.type_text(self.SEARCH_INPUT, query)

    def get_post_cards_count(self) -> int:
        return len(self.find_all(self.POST_CARDS, timeout=5))

    def click_first_edit(self):
        self.click(self.EDIT_POST_BTN)

    def click_first_block(self):
        self.click(self.BLOCK_POST_BTN)
        try:
            alert = self.driver.switch_to.alert
            alert.accept()
        except Exception:
            pass

    def get_my_posts_count(self):
        return len(self.find_all(self.MY_POSTS_CARDS, timeout=5))
