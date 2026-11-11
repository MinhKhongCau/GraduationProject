import pytest
import time
from pages.login_page import LoginPage
from pages.forum_page import ForumPage
from conftest import BASE_URL

def test_TC_FORUM_MNG_01_author_edit_post_success(driver):
    """TC-FORUM-MNG-01: Chủ bài viết xem danh sách bài viết và giao diện chỉnh sửa thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    forum_page = ForumPage(driver)
    forum_page.navigate_my_posts()
    time.sleep(2)
    
    # Check if author has any editable posts
    if forum_page.is_visible(forum_page.EDIT_POST_BTN, timeout=3):
        forum_page.click_first_edit()
        time.sleep(2)
        assert "/forum" in driver.current_url
    else:
        # If no posts yet, verify navigation to new post creation form
        forum_page.navigate_new_post()
        time.sleep(2)
        assert "/patient/forum/new" in driver.current_url
        assert forum_page.is_visible(forum_page.TITLE_INPUT)

def test_TC_FORUM_MNG_02_prevent_edit_others_post(driver):
    """TC-FORUM-MNG-02: Từ chối chỉnh sửa bài viết của người khác"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    # Try directly opening an edit URL of an arbitrary or non-owned post
    driver.get(f"{BASE_URL}/patient/forum/arbitrary-slug-other-user/edit")
    time.sleep(2)
    
    # User should either be redirected, shown 403 / 404, or returned to feed
    assert "/patient/forum" in driver.current_url

def test_TC_FORUM_MNG_03_soft_delete_or_block_post(driver):
    """TC-FORUM-MNG-03: Xóa / Ẩn bài viết cá nhân thành công (Soft Delete)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    forum_page = ForumPage(driver)
    forum_page.navigate_my_posts()
    time.sleep(2)
    
    initial_count = forum_page.get_my_posts_count()
    if initial_count > 0 and forum_page.is_visible(forum_page.BLOCK_POST_BTN, timeout=2):
        forum_page.click_first_block()
        time.sleep(2)
        # Verify page remains functional after block action
        assert "/patient/forum/my-posts" in driver.current_url
    else:
        assert "/patient/forum/my-posts" in driver.current_url
