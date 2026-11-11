from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class ForumPage(BasePage):
    # Forum list & navigation
    NEW_POST_BTN = (By.XPATH, "//a[contains(., 'New post')]")
    POST_CARDS = (By.XPATH, "//article | //div[contains(@class, 'p-5')]//h2 | //div[contains(@class, 'p-5')]//a")
    
    # Post Form locators
    CATEGORY_SELECT = (By.CSS_SELECTOR, "select#post-category, select[name='categoryId']")
    TITLE_INPUT = (By.CSS_SELECTOR, "input#post-title, input[name='title']")
    SUMMARY_INPUT = (By.CSS_SELECTOR, "input#post-summary, input[name='summary']")
    SUBMIT_BTN = (By.XPATH, "//button[@type='submit' and (contains(., 'Publish') or contains(., 'Save') or contains(., 'Saving'))]")
    ERROR_MSG = (By.CSS_SELECTOR, "p.text-danger, .error-message, [role='alert']")

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
        from selenium.webdriver.support.ui import Select
        sel = Select(self.find(self.CATEGORY_SELECT))
        # pick the second option (first real category after placeholder)
        options = sel.options
        if len(options) > 1:
            sel.select_by_index(1)

    def enter_summary(self, summary: str):
        self.type_text(self.SUMMARY_INPUT, summary)

    def click_submit(self):
        self.click(self.SUBMIT_BTN)

    def click_first_edit(self):
        self.click(self.EDIT_POST_BTN)

    def click_first_block(self):
        self.click(self.BLOCK_POST_BTN)
        # handle window.confirm
        try:
            alert = self.driver.switch_to.alert
            alert.accept()
        except Exception:
            pass

    def get_my_posts_count(self):
        return len(self.find_all(self.MY_POSTS_CARDS, timeout=5))
