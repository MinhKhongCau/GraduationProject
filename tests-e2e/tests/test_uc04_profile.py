import pytest
import time
from pages.login_page import LoginPage
from pages.profile_page import ProfilePage

def test_TC_PROF_MNG_03_view_personal_profile_success(driver):
    """TC-PROF-MNG-03: Xem profile cá nhân thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    profile_page = ProfilePage(driver)
    profile_page.navigate_patient()
    
    time.sleep(2)
    # Check that personal profile fields are retrieved and visible
    assert profile_page.is_visible(profile_page.FULL_NAME_INPUT)
    assert profile_page.is_visible(profile_page.DOB_INPUT)
    
    full_name = profile_page.get_full_name()
    assert len(full_name) > 0, "Personal full name should be loaded and displayed"

def test_TC_PROF_MNG_01_update_profile_success(driver):
    """TC-PROF-MNG-01: Cập nhật / Lưu thông tin hồ sơ thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    profile_page = ProfilePage(driver)
    profile_page.navigate_patient()
    
    time.sleep(2)
    original_name = profile_page.get_full_name()
    updated_name = "Patient Updated"
    
    profile_page.enter_full_name(updated_name)
    profile_page.click_save()
    
    # Verify success toast notification or persistent value
    time.sleep(2)
    assert profile_page.is_success_alert_present(timeout=5) or profile_page.get_full_name() == updated_name
    
    # Restore original name
    time.sleep(1)
    if original_name:
        profile_page.enter_full_name(original_name)
        profile_page.click_save()

def test_TC_PROF_MNG_02_profile_validation_empty_name(driver):
    """TC-PROF-MNG-02: Tạo/Cập nhật profile thất bại do vi phạm ràng buộc dữ liệu (Conflict/Validation)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    profile_page = ProfilePage(driver)
    profile_page.navigate_patient()
    
    time.sleep(2)
    # Enter invalid empty name
    profile_page.enter_full_name(" ")
    profile_page.click_save()
    
    error_msg = profile_page.get_error_message()
    assert "Enter your full name" in error_msg or len(error_msg) > 0
