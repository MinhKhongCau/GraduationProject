from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class NotificationPage(BasePage):
    # Potential notification elements
    NAV_NOTIFICATION_BELL = (By.CSS_SELECTOR, "button[aria-label*='notification' i], button[aria-label*='thông báo' i], a[href*='notification']")
    NOTIFICATION_BADGE = (By.CSS_SELECTOR, ".notification-badge, [aria-label*='unread' i]")
    NOTIFICATION_DROPDOWN = (By.CSS_SELECTOR, "div[role='menu'][aria-label*='notification' i], div[data-testid='notifications-panel']")
    NOTIFICATION_LIST_ITEMS = (By.CSS_SELECTOR, ".notification-item, [data-testid='notification-item']")
    MARK_ALL_READ_BTN = (By.XPATH, "//button[contains(., 'Đã đọc') or contains(., 'Mark all as read') or contains(., 'Mark as read')]")
    NOTIFICATIONS_PAGE_URL = f"{BASE_URL}/notifications"

    def navigate_notifications_page(self):
        self.driver.get(self.NOTIFICATIONS_PAGE_URL)

    def is_notification_bell_present(self, timeout=3) -> bool:
        return self.is_visible(self.NAV_NOTIFICATION_BELL, timeout=timeout)

    def open_notifications_dropdown(self):
        self.click(self.NAV_NOTIFICATION_BELL)

    def is_notifications_dropdown_open(self, timeout=3) -> bool:
        return self.is_visible(self.NOTIFICATION_DROPDOWN, timeout=timeout)

    def get_notification_items_count(self) -> int:
        return len(self.find_all(self.NOTIFICATION_LIST_ITEMS, timeout=3))

    def mark_notifications_read(self):
        self.click(self.MARK_ALL_READ_BTN)
