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
      className={`group flex flex-col ${isMine ? "items-end" : "items-start"}`}
      onMouseEnter={() => setShowReactionBar(true)}
      onMouseLeave={() => setShowReactionBar(false)}
    >
      <div className={`relative max-w-[82%] sm:max-w-[68%] ${isMine ? "pl-8" : "pr-8"}`}>
        {showReactionBar && (
          <div
            role="group"
            aria-label="Choose a reaction"
            className={`absolute -top-10 z-10 flex gap-0.5 rounded-full border border-border/70 bg-background p-1 shadow-elevated ${isMine ? "right-0" : "left-0"}`}
          >
            {QUICK_REACTIONS.map((emoji) => (
              <button
                key={emoji}
                type="button"
                onClick={() => onReact(emoji)}
                aria-label={`React ${emoji}`}
                className="inline-flex h-7 w-7 items-center justify-center rounded-full text-sm transition-transform hover:scale-125 hover:bg-surface"
              >
                {emoji}
              </button>
            ))}
          </div>
        )}

        <div
          className={`rounded-[20px] px-3.5 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap transition-transform select-none sm:select-text ${isLongPressing ? "scale-[0.97]" : "scale-100"} ${
            isMine
              ? "rounded-br-md bg-gradient-to-br from-primary to-primary-hover text-white shadow-[0_2px_8px_rgba(3,105,161,0.25)]"
              : "rounded-bl-md bg-background text-foreground shadow-card ring-1 ring-border/60"
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
              className={`flex items-center gap-2.5 py-0.5 ${isMine ? "text-white" : "text-foreground"}`}
            >
              <span className={`flex h-8 w-8 items-center justify-center rounded-full ${isMine ? "bg-white/20" : "bg-primary-soft text-primary"}`}>
                <Play className="ml-0.5 h-3.5 w-3.5 fill-current" />
              </span>
              <span className="flex items-end gap-[3px]" aria-hidden="true">
                {[8, 14, 10, 18, 12, 16, 9, 13].map((h, i) => (
                  <span key={i} style={{ height: h }} className={`w-[3px] rounded-full ${isMine ? "bg-white/60" : "bg-primary/40"}`} />
                ))}
              </span>
              <span className="text-xs tabular-nums">{formatDuration(message.duration ?? 0)}</span>
            </button>
          ) : (
            message.text
          )}
        </div>

        {reactions && Object.keys(reactions).length > 0 && (
          <div className={`-mt-1.5 flex gap-1 px-2 ${isMine ? "justify-end" : "justify-start"}`}>
            {Object.entries(reactions).map(([emoji, userIds]) => (
              <button
                key={emoji}
                type="button"
                onClick={() => onReact(emoji)}
                aria-pressed={userIds.includes(currentUserId)}
                className={`relative rounded-full border px-1.5 py-px text-[11px] shadow-card transition-colors ${
                  userIds.includes(currentUserId) ? "border-primary/40 bg-primary-soft" : "border-border bg-background hover:bg-surface"
                }`}
              >
                {emoji} {userIds.length > 1 ? userIds.length : ""}
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="mt-1 flex items-center gap-1 px-1.5">
        <span className="text-[10.5px] text-muted-foreground/80">{time}</span>
        {isMine && <CheckCheck className="h-3.5 w-3.5 text-primary" aria-label="Sent" />}
      </div>
    </div>
  );
}
