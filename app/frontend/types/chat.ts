export type ChatMessageType = "chat" | "voice";

/** Mirrors chatroom-service's dm.socket.controller.js message shape. */
export interface ChatMessage {
  id: string;
  dmId: string;
  type: ChatMessageType;
  text?: string;
  audio?: string; // base64, no `data:...;base64,` prefix
  duration?: number; // seconds
  mimeType?: string;
  fromId: string; // accountId
  toId: string; // accountId
  createdAt: number; // epoch ms
}

export interface DmHistoryPayload {
  dmId: string;
  history: ChatMessage[];
}

export interface DmReactionPayload {
  messageId: string;
  emoji: string;
  userId: string;
}

export interface DmTypingStatusPayload {
  fromId: string;
  isTyping: boolean;
}

export type ChatRole = "PATIENT" | "EXPERT";

export interface ChatContact {
  id: string; // accountId
  fullName: string;
  avatarUrl?: string;
  role: ChatRole;
}

export type ChatConnectionState = "idle" | "connecting" | "connected" | "error";

/** Per-message emoji -> reacting accountIds, maintained client-side since the
 * server only emits the toggled {messageId, emoji, userId}, not an aggregate. */
export type MessageReactions = Record<string, Record<string, string[]>>;
