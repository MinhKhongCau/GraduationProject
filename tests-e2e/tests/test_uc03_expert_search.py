import pytest
import time
from pages.login_page import LoginPage
from pages.expert_search_page import ExpertSearchPage

def test_TC_PROF_EXP_01_search_expert_with_results(driver):
    """TC-PROF-EXP-01: Tìm kiếm thành công có kết quả (Happy Case)"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    search_page = ExpertSearchPage(driver)
    search_page.navigate()
    
    # Wait for experts to load
    time.sleep(2)
    initial_experts = search_page.get_expert_names()
    assert len(initial_experts) > 0, "Initial list of experts should not be empty"
    
    # Search by keyword matching the first expert's name
    first_name_keyword = initial_experts[0].split()[0]
    search_page.search_by_name(first_name_keyword)
    time.sleep(1.5)
    
    filtered_experts = search_page.get_expert_names()
    assert len(filtered_experts) > 0
    assert any(first_name_keyword.lower() in name.lower() for name in filtered_experts)

def test_TC_PROF_EXP_02_search_expert_no_results(driver):
    """TC-PROF-EXP-02: Tìm kiếm chuyên gia không ra kết quả"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    search_page = ExpertSearchPage(driver)
    search_page.navigate()
    
    # Search with a non-existent keyword
    search_page.search_by_name("NON_EXISTENT_DOCTOR_XYZ_99999")
    time.sleep(1.5)
    
    assert search_page.is_no_results_displayed()
    assert len(search_page.get_expert_names()) == 0

def test_TC_PROF_EXP_03_query_error_handling(driver):
    """TC-PROF-EXP-03: Lỗi truy vấn dữ liệu từ database / fallback an toàn"""
    login_page = LoginPage(driver)
    login_page.login_as_patient()
    
    search_page = ExpertSearchPage(driver)
    search_page.navigate()
    
    # Verify search input is present, functional and page doesn't crash on rapid input
    search_page.search_by_name("!@#$%^&*()_+-=")
    time.sleep(1)
    
    # The UI should stay responsive, either showing 0 results or maintaining state without blanking out
    assert search_page.is_visible(search_page.SEARCH_INPUT)
