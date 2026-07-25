import pytest
from unittest.mock import MagicMock, patch
from app.layers import layer0_safety, layer1_nlu, layer2_dialogue, layer3_rag, layer4_llm, layer5_filter

def test_layer0_safety():
    # Safe message
    is_safe, emergency = layer0_safety.check_input_risk("I am feeling fine today.")
    assert is_safe is True
    assert emergency == ""

    # Unsafe message containing crisis keyword
    is_safe, emergency = layer0_safety.check_input_risk("I just want to die.")
    assert is_safe is False
    assert "psychiatric hospital hotline" in emergency.lower()

def test_layer1_nlu_text_emotion():
    assert layer1_nlu.analyze_text_emotion("I am so anxious and stressed.") == "anxiety"
    assert layer1_nlu.analyze_text_emotion("I feel sad and hopeless.") == "sadness"
    assert layer1_nlu.analyze_text_emotion("I am mad and frustrated!") == "anger"
    assert layer1_nlu.analyze_text_emotion("Just a normal day.") == "neutral"

def test_layer1_nlu_audio_emotion_fallback():
    # When model is None, it should return neutral fallback
    assert layer1_nlu.analyze_audio_emotion("dummy_path.wav") == "neutral"

def test_layer2_dialogue():
    session_id = "test-session"
    layer2_dialogue.clear_session(session_id)
    
    # Empty history
    assert layer2_dialogue.get_chat_history(session_id) == ""

    # Save turns
    layer2_dialogue.save_chat_turn(session_id, "Hello", "Hi there")
    layer2_dialogue.save_chat_turn(session_id, "How are you?", "Doing well")

    history = layer2_dialogue.get_chat_history(session_id)
    assert "User: Hello" in history
    assert "AI: Hi there" in history
    assert "User: How are you?" in history
    assert "AI: Doing well" in history

    # Clear
    layer2_dialogue.clear_session(session_id)
    assert layer2_dialogue.get_chat_history(session_id) == ""

def test_layer5_filter():
    # Safe AI response
    safe_response = "You can try deep breathing exercises."
    assert layer5_filter.filter_output(safe_response) == safe_response

    # Violating response (diagnosis/prescribing)
    violating_response = "I think you have depression. You should take 20mg of Fluoxetine."
    filtered = layer5_filter.filter_output(violating_response)
    assert filtered != violating_response
    assert "cannot provide medical diagnoses" in filtered

@patch("app.layers.layer3_rag.get_vector_db")
def test_layer3_rag(mock_get_vector_db):
    mock_db = MagicMock()
    mock_doc = MagicMock()
    mock_doc.page_content = "Deep breathing techniques help calm the nervous system."
    mock_doc.metadata = {"source_file": "breathing.pdf"}
    
    mock_db.similarity_search.return_value = [mock_doc]
    mock_get_vector_db.return_value = mock_db

    context = layer3_rag.retrieve_interventions("I feel stressed", "anxiety")
    assert "breathing.pdf" in context
    assert "Deep breathing techniques" in context
    mock_db.similarity_search.assert_called_once()
