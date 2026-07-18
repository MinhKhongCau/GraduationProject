"use client";

import { useState } from "react";
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
  const time = new Date(message.createdAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

  return (
    <div
      className={`flex flex-col ${isMine ? "items-end" : "items-start"}`}
      onMouseEnter={() => setShowReactionBar(true)}
      onMouseLeave={() => setShowReactionBar(false)}
    >
      <div
        className={`max-w-[70%] rounded-2xl px-5 py-3 text-sm leading-relaxed ${
          isMine ? "rounded-br-sm bg-primary text-white" : "rounded-bl-sm border border-border bg-background text-foreground shadow-card"
        }`}
      >
        {message.type === "voice" ? (
          <button
            type="button"
            onClick={() => message.audio && playBase64Audio(message.audio)}
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
              onClick={() => onReact(emoji)}
              className={`rounded-full border px-1.5 py-0.5 text-[11px] ${
                userIds.includes(currentUserId) ? "border-primary bg-primary-soft" : "border-border bg-surface"
              }`}
            >
              {emoji} {userIds.length}
            </button>
          ))}
        </div>
      )}

      {showReactionBar && (
        <div className="mt-1 flex gap-0.5 rounded-full border border-border bg-background px-1.5 py-0.5 shadow-card">
          {QUICK_REACTIONS.map((emoji) => (
            <button key={emoji} onClick={() => onReact(emoji)} className="rounded-full p-0.5 text-xs hover:bg-surface">
              {emoji}
            </button>
          ))}
        </div>
      )}

      <div className="mt-1.5 flex items-center gap-1 px-1">
        <span className="text-[11px] text-muted-foreground">{time}</span>
        {isMine && <CheckCheck className="h-3.5 w-3.5 text-primary" />}
      </div>
    </div>
  );
}
