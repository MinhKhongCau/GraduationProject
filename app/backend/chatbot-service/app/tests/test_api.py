
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_health_endpoint():
    response = client.get("/health")

    assert response.status_code == 200

    data = response.json()
    assert data["status"] == "ok"
    assert data["version"] == "2.0.0"
    assert data["whisper_model"] in ("loaded", "not loaded")


@patch("app.main.chat_orchestrator_stream")
def test_chat_stream_endpoint(mock_orchestrator):
    def mock_generator(*args, **kwargs):
        yield "Hello, "
        yield "this is a "
        yield "mocked response."

    mock_orchestrator.return_value = mock_generator()

    payload = {
        "message": "I feel anxious",
        "session_id": "test-session-123"
    }

    response = client.post(
        "/api/v1/chat/stream",
        json=payload
    )

    assert response.status_code == 200
    assert response.text == "Hello, this is a mocked response."

    mock_orchestrator.assert_called_once_with(
        session_id="test-session-123",
        user_message="I feel anxious"
    )


@patch("app.main.chat_orchestrator_stream")
@patch("app.main.shutil.copyfileobj")
def test_chat_voice_endpoint(mock_copyfileobj, mock_orchestrator):
    def mock_generator(*args, **kwargs):
        yield "Voice chat response"

    mock_orchestrator.return_value = mock_generator()

    audio_data = b"dummy audio bytes"

    files = {
        "audio_file": ("test.wav", audio_data, "audio/wav")
    }

    data = {
        "message": "Hello Voice",
        "session_id": "voice-session-123"
    }

    with patch("builtins.open", MagicMock()):
        response = client.post(
            "/api/v1/chat/voice",
            data=data,
            files=files
        )

    assert response.status_code == 200
    assert response.text == "Voice chat response"

    mock_copyfileobj.assert_called_once()
    mock_orchestrator.assert_called_once()
