import { CheckCheck } from "lucide-react";
import type { MockChatMessage } from "@/data";

export function MessageBubble({ message }: { message: MockChatMessage }) {
  const isPatient = message.sender === "patient";

  return (
    <div className={`flex flex-col ${isPatient ? "items-end" : "items-start"}`}>
      <div
        className={`max-w-[70%] rounded-2xl px-5 py-3 text-sm leading-relaxed ${
          isPatient ? "rounded-br-sm bg-primary text-white" : "rounded-bl-sm border border-border bg-background text-foreground shadow-card"
        }`}
      >
        {message.text}
      </div>
      <div className="mt-1.5 flex items-center gap-1 px-1">
        <span className="text-[11px] text-muted-foreground">{message.time}</span>
        {isPatient && <CheckCheck className="h-3.5 w-3.5 text-primary" />}
      </div>
    </div>
  );
}
