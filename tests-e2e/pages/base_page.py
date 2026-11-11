from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException, NoSuchElementException
import sys

class BasePage:
    def __init__(self, driver):
        self.driver = driver
        self.wait = WebDriverWait(driver, 10)

    def find(self, locator, timeout=10):
        wait = WebDriverWait(self.driver, timeout)
        return wait.until(EC.visibility_of_element_located(locator))

    def find_all(self, locator, timeout=10):
        try:
            wait = WebDriverWait(self.driver, timeout)
            return wait.until(EC.presence_of_all_elements_located(locator))
        except TimeoutException:
            return []

    def click(self, locator, timeout=10):
        wait = WebDriverWait(self.driver, timeout)
        element = wait.until(EC.element_to_be_clickable(locator))
        element.click()

    def type_text(self, locator, text, clear=True):
        element = self.find(locator)
        if clear:
            # Clear using select-all + backspace to properly notify React state
            cmd_key = Keys.COMMAND if sys.platform == "darwin" else Keys.CONTROL
            element.send_keys(cmd_key + "a")
            element.send_keys(Keys.BACKSPACE)
            element.clear()
        if text:
            element.send_keys(text)

    def get_text(self, locator, timeout=10):
        return self.find(locator, timeout).text

    def get_value(self, locator, timeout=10):
        return self.find(locator, timeout).get_attribute("value")

    def is_visible(self, locator, timeout=5):
        try:
            WebDriverWait(self.driver, timeout).until(EC.visibility_of_element_located(locator))
            return True
        except (TimeoutException, NoSuchElementException):
            return False

    def wait_for_url_contains(self, fragment, timeout=10):
        return WebDriverWait(self.driver, timeout).until(EC.url_contains(fragment))

    def scroll_into_view(self, locator):
        element = self.driver.find_element(*locator)
        self.driver.execute_script("arguments[0].scrollIntoView({block: 'center'});", element)

    def execute_script(self, script, *args):
        return self.driver.execute_script(script, *args)
