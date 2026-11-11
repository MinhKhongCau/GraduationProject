import pytest
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

BASE_URL = "https://sandbox.qmcloud.io.vn"

@pytest.fixture(scope="function")
def driver():
    options = Options()
    # Tự động cấp quyền Camera & Mic cho luồng WebRTC (UC-09, UC-10)
    options.add_argument("--use-fake-ui-for-media-stream")
    options.add_argument("--use-fake-device-for-media-stream")
    # Bỏ comment dòng dưới nếu chạy trên CI/CD (Headless)
    # options.add_argument("--headless=new")
    options.add_argument("--window-size=1920,1080")

    driver = webdriver.Chrome(options=options)
    driver.implicitly_wait(5)
    yield driver
    driver.quit()
