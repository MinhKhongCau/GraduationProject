# run_console.py
import uuid
import sys
from app.core.orchestrator import chat_orchestrator_stream

def start_console_chat():
    print("=" * 60)
    print("🧠 CLINICAL AI MENTAL WELLNESS - STREAMING DEMO")
    print("Type 'quit' or 'exit' to end the conversation.")
    print("=" * 60)

    session_id = str(uuid.uuid4())
    print(f"[System] Session ID initialized: {session_id}\n")

    while True:
        try:
            user_input = input("👤 You: ")
            
            if user_input.lower() in ['quit', 'exit']:
                print("\n[System] Shutting down. Thank you for using the system!")
                break
                
            if not user_input.strip():
                continue

            print("🌟 AI: ", end="", flush=True) # Không tự động xuống dòng
            
            # Hứng từng chữ từ generator và in ra ngay lập tức
            for chunk in chat_orchestrator_stream(session_id=session_id, user_message=user_input):
                print(chunk, end="", flush=True)
            
            print("\n") # Xuống dòng khi AI nói xong

        except KeyboardInterrupt:
            print("\n[System] Force quit detected. Goodbye!")
            break
        except Exception as e:
            print(f"\n❌ [System Error]: {e}")

if __name__ == "__main__":
    start_console_chat()