"use client";

import { useEffect, useRef, useState } from "react";
import { CheckCheck, Play } from "lucide-react";
import { playBase64Audio } from "@/lib/audio";
import type { ChatMessage } from "@/types";

const QUICK_REACTIONS = ["👍", "❤️", "😂", "🔥", "😮", "😢"];

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export interface MessageBubbleProps {
  message: ChatMessage;
  isMine: boolean;
  reactions?: Record<string, string[]>;
  currentUserId: string;
  onReact: (emoji: string) => void;
}

export function MessageBubble({ message, isMine, reactions, currentUserId, onReact }: MessageBubbleProps) {
  const [showReactionBar, setShowReactionBar] = useState(false);
  const [isLongPressing, setIsLongPressing] = useState(false);
  const longPressTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longPressTriggeredRef = useRef(false);
  const time = new Date(message.createdAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

  const clearLongPress = () => {
    if (longPressTimerRef.current) clearTimeout(longPressTimerRef.current);
    longPressTimerRef.current = null;
    setIsLongPressing(false);
  };

  useEffect(() => clearLongPress, []);

  const handleTouchStart = () => {
    longPressTriggeredRef.current = false;
    setIsLongPressing(true);
    longPressTimerRef.current = setTimeout(() => {
      longPressTriggeredRef.current = true;
      setIsLongPressing(false);
      setShowReactionBar(true);
      navigator.vibrate?.(10);
    }, 500);
  };

  const handleClick = () => {
    if (longPressTriggeredRef.current) {
      longPressTriggeredRef.current = false;
      return;
    }
    setShowReactionBar((visible) => !visible);
  };

  return (
    <div
      className={`flex flex-col ${isMine ? "items-end" : "items-start"}`}
      onMouseEnter={() => setShowReactionBar(true)}
      onMouseLeave={() => setShowReactionBar(false)}
    >
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-2.5 text-sm sm:max-w-[70%] leading-relaxed transition-transform ${isLongPressing ? "scale-[0.98]" : "scale-100"} ${
          isMine ? "rounded-br-sm bg-primary text-white" : "rounded-bl-sm border border-border bg-background text-foreground shadow-card"
        }`}
        onClick={handleClick}
        onTouchStart={handleTouchStart}
        onTouchEnd={clearLongPress}
        onTouchCancel={clearLongPress}
      >
        {message.type === "voice" ? (
          <button
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              if (message.audio) playBase64Audio(message.audio);
            }}
            className={`flex items-center gap-2 ${isMine ? "text-white" : "text-foreground"}`}
          >
            <span className={`flex h-7 w-7 items-center justify-center rounded-full ${isMine ? "bg-white/20" : "bg-primary-soft"}`}>
              <Play className="h-3.5 w-3.5" />
            </span>
            <span className="text-xs">{formatDuration(message.duration ?? 0)}</span>
          </button>
        ) : (
          message.text
        )}
      </div>

      {reactions && Object.keys(reactions).length > 0 && (
        <div className="mt-1 flex gap-1 px-1">
          {Object.entries(reactions).map(([emoji, userIds]) => (
            <button
              key={emoji}
              type="button"
              onClick={() => onReact(emoji)}
              aria-pressed={userIds.includes(currentUserId)}
              className={`rounded-full border px-2 py-0.5 text-xs transition-colors ${
                userIds.includes(currentUserId) ? "border-primary bg-primary-soft" : "border-border bg-background hover:bg-surface"
              }`}
            >
              {emoji} {userIds.length}
            </button>
          ))}
        </div>
      )}

      {showReactionBar && (
        <div className="mt-1 flex gap-0.5 rounded-full border border-border bg-background p-1 shadow-elevated" role="group" aria-label="Choose a reaction">
          {QUICK_REACTIONS.map((emoji) => (
            <button key={emoji} type="button" aria-label={`React ${emoji}`} onClick={() => onReact(emoji)} className="inline-flex h-7 w-7 items-center justify-center rounded-full text-sm transition-colors hover:bg-surface">
              {emoji}
            </button>
          ))}
        </div>
      )}

      <div className="mt-1.5 flex items-center gap-1 px-1">
        <span className="text-[11px] text-muted-foreground">{time}</span>
        {isMine && <CheckCheck className="h-3.5 w-3.5 text-primary" aria-hidden="true" />}
      </div>
    </div>
  );
}
