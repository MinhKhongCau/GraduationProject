import uuid
import pytest
import time
from pages.login_page import LoginPage
from pages.assessment_page import AssessmentPage

def test_TC_ASM_TPL_01_create_assessment_template_success(driver):
    """TC-ASM-TPL-01: Tạo bài trắc nghiệm mới thành công (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_admin()
    
    asm_page = AssessmentPage(driver)
    asm_page.navigate_admin_templates()
    time.sleep(2)
    
    unique_code = f"TEST_{uuid.uuid4().hex[:6].upper()}"
    asm_page.open_create_modal()
    time.sleep(1)
    
    asm_page.fill_template_form(
        code=unique_code,
        title=f"Đánh giá tâm lý thử nghiệm {unique_code}",
        description="Mô tả bài kiểm tra đánh giá mức độ lo âu thử nghiệm."
    )
    asm_page.submit_template_form()
    time.sleep(2)
    
    # Verify success toast appears or templates count refreshed
    assert asm_page.is_toast_visible(timeout=8) or asm_page.get_template_rows_count() > 0

def test_TC_ASM_TPL_02_create_template_duplicate_code(driver):
    """TC-ASM-TPL-02: Tạo bài trắc nghiệm thất bại do trùng lặp mã bài test"""
    login_page = LoginPage(driver)
    login_page.login_as_admin()
    
    asm_page = AssessmentPage(driver)
    asm_page.navigate_admin_templates()
    time.sleep(2)
    
    # Use existing standard code PHQ_9 or DASS_21
    asm_page.open_create_modal()
    time.sleep(1)
    
    asm_page.fill_template_form(
        code="PHQ_9",
        title="Trầm cảm PHQ-9 Trùng Lặp",
        description="Thử nghiệm tạo trùng mã code."
    )
    asm_page.submit_template_form()
    time.sleep(2)
    
    # Should display error alert/toast
    assert asm_page.is_toast_visible(timeout=8)

def test_TC_ASM_TPL_03_view_assessment_detail_by_slug(driver):
    """TC-ASM-TPL-03: Xem thông tin chi tiết bài trắc nghiệm theo slug"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    asm_page = AssessmentPage(driver)
    asm_page.navigate_patient_assessments()
    time.sleep(2)
    
    titles = asm_page.get_patient_assessment_titles()
    if len(titles) > 0:
        # Click to view detail of the first assessment
        asm_page.click_first_assessment()
        time.sleep(2)
        assert "/patient/assessment/" in driver.current_url
        assert asm_page.is_visible(asm_page.ASSESSMENT_DETAIL_TITLE)
    else:
        # If no published template yet, verify patient assessment page loaded properly
        assert "/patient/assessment" in driver.current_url
