from selenium.webdriver.common.by import By
from pages.base_page import BasePage
from conftest import BASE_URL

class ChatPage(BasePage):
    # Navigation
    PATIENT_MESSAGES_URL = f"{BASE_URL}/patient/messages"
    EXPERT_MESSAGES_URL = f"{BASE_URL}/expert/messages"

    # Contact List
    CONTACT_ITEMS = (By.XPATH, "//div[contains(@class, 'overflow-y-auto')]//button[contains(@class, 'w-full')]")
    SEARCH_INPUT = (By.CSS_SELECTOR, "input[aria-label='Search experts'], input[placeholder*='Search']")
    ACTIVE_CONTACT = (By.CSS_SELECTOR, "button[aria-current='true']")
    EMPTY_CONVERSATIONS_MSG = (By.XPATH, "//*[contains(text(), 'No conversations yet') or contains(text(), 'Select a conversation')]")

    # Chat Window
    MESSAGE_INPUT = (By.CSS_SELECTOR, "input[aria-label='Type your message'], input[placeholder*='Type your message']")
    SEND_BUTTON = (By.CSS_SELECTOR, "button[aria-label='Send message']")
    VIDEO_CALL_BUTTON = (By.CSS_SELECTOR, "button[aria-label='Video Call']")
    MESSAGE_BUBBLES = (By.XPATH, "//div[contains(@class, 'space-y-3')]//p")
    BACK_TO_CONTACTS = (By.CSS_SELECTOR, "button[aria-label='Back to contacts']")

    def navigate_patient_messages(self):
        self.driver.get(self.PATIENT_MESSAGES_URL)

    def navigate_expert_messages(self):
        self.driver.get(self.EXPERT_MESSAGES_URL)

    def navigate_meet(self, meet_id: str):
        self.driver.get(f"{BASE_URL}/meet/{meet_id}")

    def get_contacts_count(self) -> int:
        return len(self.find_all(self.CONTACT_ITEMS, timeout=5))

    def select_first_contact(self):
        contacts = self.find_all(self.CONTACT_ITEMS, timeout=5)
        if contacts:
            contacts[0].click()

    def is_message_input_visible(self) -> bool:
        return self.is_visible(self.MESSAGE_INPUT, timeout=5)

    def enter_message(self, text: str):
        self.type_text(self.MESSAGE_INPUT, text)

    def is_send_button_enabled(self) -> bool:
        el = self.find(self.SEND_BUTTON, timeout=3)
        return el.is_enabled() if el else False

    def click_send(self):
        self.click(self.SEND_BUTTON)

    def is_empty_state_visible(self) -> bool:
        return self.is_visible(self.EMPTY_CONVERSATIONS_MSG, timeout=5)

    def is_video_call_button_visible(self) -> bool:
        return self.is_visible(self.VIDEO_CALL_BUTTON, timeout=5)
