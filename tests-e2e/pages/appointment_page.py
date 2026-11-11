from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class AppointmentPage(BasePage):
    # Patient bookings locators
    BOOKINGS_TITLE = (By.XPATH, "//h1[contains(., 'Bookings') or contains(., 'Lịch hẹn')]")
    APPOINTMENT_ROWS = (By.XPATH, "//table//tbody//tr")
    EMPTY_STATE = (By.XPATH, "//p[contains(., 'No bookings') or contains(., 'Không có lịch hẹn') or contains(., 'No appointments')] | //div[contains(., 'No bookings')]")
    
    # Filter links / tabs
    STATUS_FILTER_ALL = (By.XPATH, "//nav//a[contains(., 'All') or contains(., 'Tất cả')] | //button[contains(., 'All')]")
    STATUS_FILTER_PENDING = (By.XPATH, "//nav//a[contains(., 'Pending') or contains(., 'Chờ')] | //button[contains(., 'Pending')]")
    STATUS_FILTER_CONFIRMED = (By.XPATH, "//nav//a[contains(., 'Confirmed') or contains(., 'Đã xác nhận')] | //button[contains(., 'Confirmed')]")
    
    # Action dropdown & Modal
    ROW_ACTION_TRIGGER = (By.XPATH, "(//table//tbody//tr//button[contains(., 'Thao tác') or contains(@aria-label, 'Thao tác')])[1]")
    ACTION_DETAIL_ITEM = (By.XPATH, "//div[@role='menuitem' and contains(., 'Chi tiết')] | //button[contains(., 'Chi tiết')]")
    DETAIL_MODAL = (By.XPATH, "//div[@role='dialog']")
    MODAL_CLOSE_BTN = (By.XPATH, "//div[@role='dialog']//button[contains(@aria-label, 'Close') or contains(., 'Đóng')]")

    def navigate_patient_bookings(self):
        self.driver.get(f"{BASE_URL}/patient/my-bookings")

    def navigate_expert_appointments(self):
        self.driver.get(f"{BASE_URL}/expert/appointments")

    def get_appointment_rows_count(self):
        rows = self.find_all(self.APPOINTMENT_ROWS, timeout=5)
        return len(rows)

    def is_empty_state_visible(self):
        return self.is_visible(self.EMPTY_STATE, timeout=5)

    def open_first_appointment_detail(self):
        # Click action button on first row
        self.click(self.ROW_ACTION_TRIGGER)
        # Click "Chi tiết" from menu
        self.click(self.ACTION_DETAIL_ITEM)

    def is_detail_modal_visible(self):
        return self.is_visible(self.DETAIL_MODAL, timeout=5)

    def close_detail_modal(self):
        self.click(self.MODAL_CLOSE_BTN)
