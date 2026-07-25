import pytest
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_health_endpoint():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}

@patch("app.main.chat_orchestrator_stream")
def test_chat_stream_endpoint(mock_orchestrator):
    # Mock generator response
    async def mock_generator(*args, **kwargs):
        yield b"Hello, "
        yield b"this is a "
        yield b"mocked response."

    mock_orchestrator.return_value = mock_generator()

    payload = {
        "message": "I feel anxious",
        "session_id": "test-session-123"
    }
    response = client.post("/api/v1/chat/stream", json=payload)
    
    assert response.status_code == 200
    assert response.text == "Hello, this is a mocked response."
    mock_orchestrator.assert_called_once_with(session_id="test-session-123", user_message="I feel anxious")

@patch("app.main.chat_orchestrator_stream")
@patch("shutil.copyfileobj")
def test_chat_voice_endpoint(mock_copyfileobj, mock_orchestrator):
    async def mock_generator(*args, **kwargs):
        yield b"Voice chat response"

    mock_orchestrator.return_value = mock_generator()

    # Create dummy file content
    audio_data = b"dummy audio bytes"
    files = {
        "audio_file": ("test.wav", audio_data, "audio/wav")
    }
    data = {
        "message": "Hello Voice",
        "session_id": "voice-session-123"
    }

    # Mock open to avoid actual file write during test
    with patch("builtins.open", MagicMock()):
        response = client.post("/api/v1/chat/voice", data=data, files=files)

    assert response.status_code == 200
    assert response.text == "Voice chat response"
    mock_orchestrator.assert_called_once()
