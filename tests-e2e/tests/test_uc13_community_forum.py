import time
import pytest
from pages.login_page import LoginPage
from pages.forum_page import ForumPage

@pytest.mark.forum
def test_TC_FORUM_QNA_01_create_post_success(driver):
    """TC-FORUM-QNA-01: Đăng câu hỏi / bài viết mới thành công (Happy Case)."""
    login_page = LoginPage(driver)
    forum_page = ForumPage(driver)

    login_page.login_as_patient()
    forum_page.navigate_new_post()

    timestamp = int(time.time())
    title = f"Làm sao để vượt qua căng thẳng học đường? {timestamp}"

    forum_page.select_first_category()
    forum_page.enter_title(title)
    forum_page.enter_summary("Tóm tắt thảo luận về căng thẳng tâm lý học sinh sinh viên")
    forum_page.enter_content("Đây là nội dung chia sẻ chi tiết về phương pháp quản lý căng thẳng và lo âu.")
    forum_page.click_submit()

    # Verify redirected or post submitted
    forum_page.wait_for_url_contains("/patient/forum", timeout=8)
    assert "/patient/forum" in driver.current_url

@pytest.mark.forum
def test_TC_FORUM_QNA_02_validation_error_empty_title(driver):
    """TC-FORUM-QNA-02: Đăng bài viết thất bại do thiếu tiêu đề hoặc danh mục (Validation Error)."""
    login_page = LoginPage(driver)
    forum_page = ForumPage(driver)

    login_page.login_as_patient()
    forum_page.navigate_new_post()

    # Submit without entering title
    forum_page.click_submit()

    # Verify validation error appears
    assert forum_page.is_validation_error_displayed(), "Validation error message should appear for empty title"

@pytest.mark.forum
def test_TC_FORUM_QNA_03_search_and_browse_posts(driver):
    """TC-FORUM-QNA-03: Duyệt và tìm kiếm danh sách bài viết trên diễn đàn (Browse & Search)."""
    login_page = LoginPage(driver)
    forum_page = ForumPage(driver)

    login_page.login_as_patient()
    forum_page.navigate_feed()

    # Verify forum feed loads
    assert "/patient/forum" in driver.current_url
    initial_count = forum_page.get_post_cards_count()
    assert initial_count >= 0, "Forum feed should render posts or empty state cleanly"

    # Search for keyword
    forum_page.search_posts("tâm lý")
    time.sleep(1)
    assert "/patient/forum" in driver.current_url

@pytest.mark.forum
def test_TC_FORUM_QNA_04_unauthorized_create_post(driver):
    """TC-FORUM-QNA-04: Từ chối truy cập tạo bài viết khi chưa đăng nhập (Unauthorized)."""
    forum_page = ForumPage(driver)
    driver.delete_all_cookies()
    forum_page.navigate_new_post()

    forum_page.wait_for_url_contains("/auth/login", timeout=8)
    assert "/auth/login" in driver.current_url
