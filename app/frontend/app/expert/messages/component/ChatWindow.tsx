"use client";

import Image from "next/image";
import { useRef, useState } from "react";
import { MoreVertical, SendHorizonal, Circle } from "lucide-react";
import { MessageBubble } from "./MessageBubble";
import { VoiceRecorderButton } from "./VoiceRecorderButton";
import type { ChatContact, ChatMessage } from "@/types";

const TYPING_IDLE_MS = 2000;

export interface ChatWindowProps {
  contact: ChatContact;
  messages: ChatMessage[];
  myId: string;
  isTyping: boolean;
  online: boolean;
  reactionsByMessageId: Record<string, Record<string, string[]>>;
  onSend: (text: string) => void;
  onSendVoice: (blob: Blob, duration: number) => void;
  onReact: (messageId: string, emoji: string) => void;
  onTyping: (isTyping: boolean) => void;
}

export function ChatWindow({
  contact,
  messages,
  myId,
  isTyping,
  online,
  reactionsByMessageId,
  onSend,
  onSendVoice,
  onReact,
  onTyping,
}: ChatWindowProps) {
  const [text, setText] = useState("");
  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const notifyTyping = () => {
    onTyping(true);
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    typingTimeoutRef.current = setTimeout(() => onTyping(false), TYPING_IDLE_MS);
  };

  const send = () => {
    const trimmed = text.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setText("");
    onTyping(false);
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
  };

  return (
    <div className="flex min-w-0 flex-1 flex-col bg-background">
      <div className="flex items-center justify-between border-b border-border p-4">
        <div className="flex items-center gap-3">
          <div className="relative">
            <Image
              src={contact.avatarUrl ?? "https://i.pravatar.cc/150?u=" + contact.id}
              alt={contact.fullName}
              width={40}
              height={40}
              unoptimized
              className="h-10 w-10 rounded-full object-cover"
            />
            {online && <span className="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border-2 border-background bg-success" />}
          </div>
          <div>
            <h3 className="text-sm font-bold text-foreground">{contact.fullName}</h3>
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              {isTyping ? (
                <span className="font-medium text-primary">Typing…</span>
              ) : online ? (
                <>
                  <Circle className="h-1 w-1 fill-border text-border" />
                  <span className="font-medium text-primary">Active now</span>
                </>
              ) : (
                <span>Offline</span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2 text-muted-foreground">
          <button className="rounded-full p-2 transition-colors hover:bg-surface">
            <MoreVertical className="h-5 w-5" />
          </button>
        </div>
      </div>

      <div className="flex-1 space-y-6 overflow-y-auto bg-surface/30 p-6">
        <div className="flex justify-center">
          <span className="rounded-full bg-surface px-3 py-1 text-[11px] font-medium text-muted-foreground">Today</span>
        </div>
        {messages.map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            isMine={message.fromId === myId}
            reactions={reactionsByMessageId[message.id]}
            currentUserId={myId}
            onReact={(emoji) => onReact(message.id, emoji)}
          />
        ))}
      </div>

      <div className="border-t border-border bg-background p-4">
        <div className="flex items-center gap-2 rounded-2xl border border-border bg-surface p-2">
          <VoiceRecorderButton onSend={onSendVoice} />
          <input
            type="text"
            value={text}
            onChange={(e) => {
              setText(e.target.value);
              notifyTyping();
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") send();
            }}
            onBlur={() => onTyping(false)}
            placeholder="Type your message..."
            className="flex-1 border-none bg-transparent px-2 text-sm outline-none"
          />
          <button
            onClick={send}
            className="shrink-0 rounded-full p-2 text-muted-foreground transition-colors hover:bg-primary-soft hover:text-primary"
          >
            <SendHorizonal className="h-5 w-5" />
          </button>
        </div>
        <p className="mt-3 text-center text-[10px] font-medium italic text-muted-foreground">
          Your consultation content is strictly confidential and protected by MindCare.
        </p>
      </div>
    </div>
  );
}
