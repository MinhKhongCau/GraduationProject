"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { getChatSocket } from "@/lib/chatSocket";
import { blobToBase64 } from "@/lib/audio";
import { useAuthContext } from "./AuthContext";
import type {
  ChatConnectionState,
  ChatMessage,
  DmHistoryPayload,
  DmReactionPayload,
  DmTypingStatusPayload,
  MessageReactions,
} from "@/types";

const TYPING_TIMEOUT_MS = 3000;

interface ChatContextValue {
  connectionState: ChatConnectionState;
  messagesByDmId: Record<string, ChatMessage[]>;
  reactionsByMessageId: MessageReactions;
  typingByContactId: Record<string, boolean>;
  onlineByContactId: Record<string, boolean>;
  fetchHistory: (contactId: string, page?: number, limit?: number) => void;
  sendText: (contactId: string, text: string) => void;
  sendVoice: (contactId: string, blob: Blob, duration: number) => Promise<void>;
  react: (contactId: string, messageId: string, emoji: string) => void;
  setTyping: (contactId: string, isTyping: boolean) => void;
  queryPresence: (contactIds: string[]) => void;
}

const ChatContext = createContext<ChatContextValue | null>(null);

export function ChatProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuthContext();
  const socket = useMemo(() => getChatSocket(), []);

  const [connectionState, setConnectionState] = useState<ChatConnectionState>("idle");
  const [messagesByDmId, setMessagesByDmId] = useState<Record<string, ChatMessage[]>>({});
  const [reactionsByMessageId, setReactionsByMessageId] = useState<MessageReactions>({});
  const [typingByContactId, setTypingByContactId] = useState<Record<string, boolean>>({});
  const [onlineByContactId, setOnlineByContactId] = useState<Record<string, boolean>>({});

  const typingTimeoutsRef = useRef<Record<string, ReturnType<typeof setTimeout>>>({});

  useEffect(() => {
    if (!isAuthenticated) {
      socket.disconnect();
      setConnectionState("idle");
      return;
    }

    setConnectionState("connecting");
    socket.connect();

    const onConnect = () => setConnectionState("connected");
    const onDisconnect = () => setConnectionState("idle");
    const onConnectError = () => setConnectionState("error");

    const onDmMessage = (msg: ChatMessage) => {
      setMessagesByDmId((prev) => {
        const existing = prev[msg.dmId] ?? [];
        if (existing.some((m) => m.id === msg.id)) return prev;
        return { ...prev, [msg.dmId]: [...existing, msg] };
      });
    };

    const onDmHistory = ({ dmId, history, page = 1 }: DmHistoryPayload) => {
      setMessagesByDmId((prev) => {
        const existing = prev[dmId] ?? [];
        if (page === 1) {
          return { ...prev, [dmId]: history };
        } else {
          // Prepend older history to existing messages, avoiding duplicates
          const existingIds = new Set(existing.map((m) => m.id));
          const filteredHistory = history.filter((m) => !existingIds.has(m.id));
          return { ...prev, [dmId]: [...filteredHistory, ...existing] };
        }
      });
    };

    const onDmReaction = ({ messageId, emoji, userId }: DmReactionPayload) => {
      setReactionsByMessageId((prev) => {
        const forMessage = { ...(prev[messageId] ?? {}) };
        const users = forMessage[emoji] ?? [];
        forMessage[emoji] = users.includes(userId)
          ? users.filter((id) => id !== userId)
          : [...users, userId];
        if (forMessage[emoji].length === 0) delete forMessage[emoji];
        return { ...prev, [messageId]: forMessage };
      });
    };

    const onDmTypingStatus = ({ fromId, isTyping }: DmTypingStatusPayload) => {
      setTypingByContactId((prev) => ({ ...prev, [fromId]: isTyping }));
      const timeouts = typingTimeoutsRef.current;
      if (timeouts[fromId]) clearTimeout(timeouts[fromId]);
      if (isTyping) {
        timeouts[fromId] = setTimeout(() => {
          setTypingByContactId((prev) => ({ ...prev, [fromId]: false }));
        }, TYPING_TIMEOUT_MS);
      }
    };

    const onPresenceStatus = (status: Record<string, boolean>) => {
      setOnlineByContactId((prev) => ({ ...prev, ...status }));
    };

    socket.on("connect", onConnect);
    socket.on("disconnect", onDisconnect);
    socket.on("connect_error", onConnectError);
    socket.on("dm:message", onDmMessage);
    socket.on("dm:history", onDmHistory);
    socket.on("dm:reaction", onDmReaction);
    socket.on("dm:typing:status", onDmTypingStatus);
    socket.on("presence:status", onPresenceStatus);

    return () => {
      socket.off("connect", onConnect);
      socket.off("disconnect", onDisconnect);
      socket.off("connect_error", onConnectError);
      socket.off("dm:message", onDmMessage);
      socket.off("dm:history", onDmHistory);
      socket.off("dm:reaction", onDmReaction);
      socket.off("dm:typing:status", onDmTypingStatus);
      socket.off("presence:status", onPresenceStatus);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, socket]);

  const fetchHistory = useCallback(
    (contactId: string, page = 1, limit = 20) => socket.emit("dm:history", { toUser: contactId, page, limit }),
    [socket]
  );

  const sendText = useCallback(
    (contactId: string, text: string) => {
      const trimmed = text.trim();
      if (!trimmed) return;
      socket.emit("dm:send", { toUser: contactId, text: trimmed });
    },
    [socket]
  );

  const sendVoice = useCallback(
    async (contactId: string, blob: Blob, duration: number) => {
      const audio = await blobToBase64(blob);
      socket.emit("dm:send:voice", { toUser: contactId, audio, duration, mimeType: blob.type });
    },
    [socket]
  );

  const react = useCallback(
    (contactId: string, messageId: string, emoji: string) =>
      socket.emit("dm:react", { toUser: contactId, messageId, emoji }),
    [socket]
  );

  const setTyping = useCallback(
    (contactId: string, isTyping: boolean) =>
      socket.emit("dm:typing", { toUser: contactId, isTyping }),
    [socket]
  );

  const queryPresence = useCallback(
    (contactIds: string[]) => {
      if (contactIds.length > 0) socket.emit("presence:query", { userIds: contactIds });
    },
    [socket]
  );

  const value = useMemo(
    () => ({
      connectionState,
      messagesByDmId,
      reactionsByMessageId,
      typingByContactId,
      onlineByContactId,
      fetchHistory,
      sendText,
      sendVoice,
      react,
      setTyping,
      queryPresence,
    }),
    [
      connectionState,
      messagesByDmId,
      reactionsByMessageId,
      typingByContactId,
      onlineByContactId,
      fetchHistory,
      sendText,
      sendVoice,
      react,
      setTyping,
      queryPresence,
    ]
  );

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChatContext(): ChatContextValue {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error("useChatContext must be used within a ChatProvider");
  }
  return context;
}
