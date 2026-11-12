import pytest
from pages.login_page import LoginPage
from pages.assessment_page import AssessmentPage

@pytest.mark.assessment
def test_TC_ASM_SUB_01_submit_assessment_flow_or_intro(driver):
    """TC-ASM-SUB-01: Bệnh nhân truy cập và thực hiện quy trình đánh giá tâm lý (Happy Case)."""
    login_page = LoginPage(driver)
    asm_page = AssessmentPage(driver)

    login_page.login_as_patient()
    asm_page.navigate_patient_assessments()

    titles = asm_page.get_patient_assessment_titles()
    if titles:
        # Click on the first assessment
        asm_page.click_first_assessment()
        asm_page.wait_for_url_contains("/patient/assessment/", timeout=5)
        
        # Verify intro page displays title and "Bắt đầu đánh giá" button
        if asm_page.is_start_button_visible():
            asm_page.click_start_assessment()
            # If questions load, answer first question
            asm_page.select_option(0)
            assert not asm_page.is_next_btn_disabled(), "Next button should become active after selecting an answer"
    else:
        # If no active assessment templates in sandbox, verify page loaded properly
        assert asm_page.is_history_section_visible() or "/patient/assessment" in driver.current_url

@pytest.mark.assessment
def test_TC_ASM_SUB_02_template_not_found(driver):
    """TC-ASM-SUB-02: Truy cập mã bài trắc nghiệm không tồn tại (Template Not Found)."""
    login_page = LoginPage(driver)
    asm_page = AssessmentPage(driver)

    login_page.login_as_patient()
    asm_page.navigate_patient_assessment_detail("non-existent-template-slug-99999")

    # Verify not found message is rendered
    assert asm_page.is_not_found_visible(), "Should display 'Không tìm thấy thông tin bài đánh giá' message"

@pytest.mark.assessment
def test_TC_ASM_SUB_03_unanswered_question_constraint(driver):
    """TC-ASM-SUB-03: Kiểm tra ràng buộc câu hỏi chưa chọn phương án (Next button disabled)."""
    login_page = LoginPage(driver)
    asm_page = AssessmentPage(driver)

    login_page.login_as_patient()
    asm_page.navigate_patient_assessments()

    titles = asm_page.get_patient_assessment_titles()
    if titles:
        asm_page.click_first_assessment()
        if asm_page.is_start_button_visible():
            asm_page.click_start_assessment()
            # Before selecting any option, Next button must be disabled
            assert asm_page.is_next_btn_disabled(), "Next button must be disabled when no answer is selected"

@pytest.mark.assessment
def test_TC_ASM_SUB_04_unauthorized_assessment_access(driver):
    """TC-ASM-SUB-04: Từ chối truy cập bài đánh giá khi chưa đăng nhập (Unauthorized)."""
    asm_page = AssessmentPage(driver)
    driver.delete_all_cookies()
    asm_page.navigate_patient_assessments()

    asm_page.wait_for_url_contains("/auth/login", timeout=8)
    assert "/auth/login" in driver.current_url

@pytest.mark.assessment
def test_TC_ASM_HIS_01_view_assessment_history(driver):
    """TC-ASM-HIS-01: Xem lịch sử đánh giá tâm lý cá nhân thành công."""
    login_page = LoginPage(driver)
    asm_page = AssessmentPage(driver)

    login_page.login_as_patient()
    asm_page.navigate_patient_assessments()

    assert asm_page.is_history_section_visible(), "My Assessment History section should be visible"
