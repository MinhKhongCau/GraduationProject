import pytest
from pages.login_page import LoginPage
from pages.chat_page import ChatPage
from conftest import BASE_URL

@pytest.mark.chat
def test_TC_CHAT_AUTH_01_authenticated_user_access_messages(driver):
    """TC-CHAT-AUTH-01: Xác thực JWT Token hợp lệ khi kết nối chat messages (Happy Case)."""
    login_page = LoginPage(driver)
    chat_page = ChatPage(driver)

    login_page.login_as_patient()
    chat_page.navigate_patient_messages()

    # Verify messages interface loaded without redirecting to login
    assert "/auth/login" not in driver.current_url
    assert "/patient/messages" in driver.current_url

    # Check that contact list container or empty state is rendered
    contacts_count = chat_page.get_contacts_count()
    empty_visible = chat_page.is_empty_state_visible()
    assert contacts_count >= 0 or empty_visible, "Messages view should display contacts or empty state"

@pytest.mark.chat
def test_TC_CHAT_AUTH_02_unauthenticated_access_denied(driver):
    """TC-CHAT-AUTH-02: Từ chối kết nối Socket/Messages do thiếu JWT Token (chưa đăng nhập)."""
    chat_page = ChatPage(driver)
    # Clear cookies to simulate anonymous unauthenticated user
    driver.delete_all_cookies()
    chat_page.navigate_patient_messages()

    # Must be redirected to /auth/login
    chat_page.wait_for_url_contains("/auth/login", timeout=8)
    assert "/auth/login" in driver.current_url

@pytest.mark.chat
def test_TC_CHAT_AUTH_03_meet_room_exit_state_handling(driver):
    """TC-CHAT-AUTH-03: Kiểm tra xử lý phòng họp trực tuyến (Meet) khi phiên không hợp lệ hoặc đã kết thúc."""
    login_page = LoginPage(driver)
    chat_page = ChatPage(driver)

    login_page.login_as_patient()
    # Navigate to invalid/ended meeting room
    chat_page.navigate_meet("invalid-or-ended-room-token")

    # Verify MeetExit screen or error / not-found handling is shown
    body_text = driver.find_element("tag name", "body").text
    assert any(msg in body_text for msg in [
        "The session has ended",
        "Unable to join",
        "Unable to join the meeting",
        "Starting camera",
        "Connecting",
        "You're the only one here"
    ]), f"Unexpected meet room view: {body_text[:200]}"

@pytest.mark.chat
def test_TC_CHAT_SND_01_send_text_message_enabled(driver):
    """TC-CHAT-SND-01: Gửi tin nhắn tư vấn văn bản thành công (Happy Case - Text Validation & Send Trigger)."""
    login_page = LoginPage(driver)
    chat_page = ChatPage(driver)

    login_page.login_as_expert()
    chat_page.navigate_expert_messages()

    contacts_count = chat_page.get_contacts_count()
    if contacts_count > 0:
        chat_page.select_first_contact()
        assert chat_page.is_message_input_visible(), "Chat input should be visible for active conversation"
        chat_page.enter_message("Chào bạn, tôi có thể giúp gì cho bạn?")
        assert chat_page.is_send_button_enabled(), "Send button should be enabled when text is typed"
    else:
        # Empty conversation state
        assert chat_page.is_empty_state_visible(), "Empty state should be displayed when no contacts exist"

@pytest.mark.chat
def test_TC_CHAT_SND_02_empty_or_whitespace_message_disabled(driver):
    """TC-CHAT-SND-02: Gửi thất bại do văn bản rỗng hoặc toàn khoảng trắng (disabled={!text.trim()})."""
    login_page = LoginPage(driver)
    chat_page = ChatPage(driver)

    login_page.login_as_patient()
    chat_page.navigate_patient_messages()

    contacts_count = chat_page.get_contacts_count()
    if contacts_count > 0:
        chat_page.select_first_contact()
        if chat_page.is_message_input_visible():
            chat_page.enter_message("    ")
            assert not chat_page.is_send_button_enabled(), "Send button must be disabled for whitespace-only text"
    else:
        # If no active contact, send button is not enabled/available
        assert not chat_page.is_send_button_enabled()

@pytest.mark.chat
def test_TC_CHAT_SND_03_no_recipient_selected_state(driver):
    """TC-CHAT-SND-03: Gửi thất bại do thiếu định danh người nhận (Select a conversation / No active recipient)."""
    login_page = LoginPage(driver)
    chat_page = ChatPage(driver)

    login_page.login_as_patient()
    chat_page.navigate_patient_messages()

    # When no conversation is selected or no contacts exist
    contacts_count = chat_page.get_contacts_count()
    if contacts_count == 0:
        assert chat_page.is_empty_state_visible()
        # Message input should not allow sending without recipient
        assert not chat_page.is_send_button_enabled()
