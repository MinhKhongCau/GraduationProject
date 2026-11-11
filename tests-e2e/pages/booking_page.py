import time
from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class BookingPage(BasePage):
    # Step headers / accordion buttons
    SPECIALIZATION_SECTION_BTN = (By.XPATH, "//section//button[contains(., 'Specialization')] | (//section/button)[1]")
    EXPERT_SECTION_BTN = (By.XPATH, "//section//button[contains(., 'Expert')] | (//section/button)[2]")
    SCHEDULE_SECTION_BTN = (By.XPATH, "//section//button[contains(., 'Date & time') or contains(., 'Date and time')] | (//section/button)[3]")
    PROFILE_SECTION_BTN = (By.XPATH, "//section//button[contains(., 'Patient profile')] | (//section/button)[4]")
    
    # Items inside sections
    SPECIALIZATION_CHIPS = (By.XPATH, "//button[contains(@class, 'rounded-full') and not(contains(., 'Show all'))]")
    EXPERT_CHIPS = (By.XPATH, "//div[contains(@class, 'grid')]//button")
    TIME_SLOT_BTNS = (By.XPATH, "//button[contains(@class, 'rounded-full') and contains(., ':')]")
    PATIENT_RECORD_CARDS = (By.XPATH, "//div[contains(@class, 'cursor-pointer') or contains(@class, 'rounded-xl')]//h4")
    
    # Navigation buttons
    NEXT_BTN = (By.XPATH, "//button[contains(., 'Next') or contains(., 'Holding your slot')]")
    CONFIRM_BOOKING_BTN = (By.XPATH, "//button[contains(., 'Confirm booking') or contains(., 'Booking...')]")
    BACK_BTN = (By.XPATH, "//button[contains(., 'Back')]")
    PICK_ANOTHER_SLOT_BTN = (By.XPATH, "//button[contains(., 'Pick another slot')]")
    
    # Status / Confirmation elements
    CONFIRMATION_CARD = (By.XPATH, "//h2[contains(., 'Review and confirm') or contains(., 'Booking')] | //div[contains(@class, 'grid')]")
    HOLD_EXPIRED_ALERT = (By.XPATH, "//div[contains(@class, 'bg-danger-soft') or contains(@class, 'border-danger')]")
    PAYMENT_GATEWAY_CONTAINER = (By.XPATH, "//h2[contains(., 'Payment') or contains(., 'VNPay')] | //button[contains(., 'Pay')]")

    def navigate(self, expert_id: str = None):
        if expert_id:
            self.driver.get(f"{BASE_URL}/patient/book-appointment?expertId={expert_id}")
        else:
            self.driver.get(f"{BASE_URL}/patient/book-appointment")

    def open_section(self, section_name: str):
        if "spec" in section_name.lower():
            self.click(self.SPECIALIZATION_SECTION_BTN)
        elif "exp" in section_name.lower():
            self.click(self.EXPERT_SECTION_BTN)
        elif "sched" in section_name.lower() or "date" in section_name.lower():
            self.click(self.SCHEDULE_SECTION_BTN)
        elif "prof" in section_name.lower():
            self.click(self.PROFILE_SECTION_BTN)

    def select_first_specialization(self):
        chips = self.find_all(self.SPECIALIZATION_CHIPS, timeout=5)
        if chips:
            self.execute_script("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", chips[0])

    def select_first_expert(self):
        self.open_section("expert")
        time.sleep(1)
        chips = self.find_all(self.EXPERT_CHIPS, timeout=5)
        if chips:
            self.execute_script("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", chips[0])

    def select_first_time_slot(self):
        self.open_section("schedule")
        time.sleep(1)
        slots = self.find_all(self.TIME_SLOT_BTNS, timeout=5)
        if slots:
            self.execute_script("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", slots[0])

    def select_first_patient_record(self):
        self.open_section("profile")
        time.sleep(1)
        records = self.find_all(self.PATIENT_RECORD_CARDS, timeout=5)
        if records:
            self.execute_script("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", records[0])

    def click_next(self):
        self.click(self.NEXT_BTN)

    def click_confirm_booking(self):
        self.click(self.CONFIRM_BOOKING_BTN)

    def is_hold_expired_shown(self):
        return self.is_visible(self.HOLD_EXPIRED_ALERT, timeout=5)

    def is_confirm_step_visible(self):
        return self.is_visible(self.CONFIRM_BOOKING_BTN, timeout=8)
