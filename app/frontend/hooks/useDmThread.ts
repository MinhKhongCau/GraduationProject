"use client";

import { useEffect, useMemo } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useChatContext } from "@/context/ChatContext";
import { makeDmId } from "@/lib/chatSocket";

/** Thin per-contact view over ChatContext's global socket state. */
export function useDmThread(contactId: string | null) {
  const { user } = useAuthContext();
  const chat = useChatContext();

  const dmId = contactId && user ? makeDmId(user.id, contactId) : null;

  useEffect(() => {
    if (contactId) chat.fetchHistory(contactId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [contactId]);

  const messages = useMemo(() => (dmId ? chat.messagesByDmId[dmId] ?? [] : []), [dmId, chat.messagesByDmId]);
  const isTyping = contactId ? !!chat.typingByContactId[contactId] : false;
  const online = contactId ? !!chat.onlineByContactId[contactId] : false;

  return {
    messages,
    isTyping,
    online,
    reactionsByMessageId: chat.reactionsByMessageId,
    sendText: (text: string) => contactId && chat.sendText(contactId, text),
    sendVoice: (blob: Blob, duration: number) => contactId && chat.sendVoice(contactId, blob, duration),
    react: (messageId: string, emoji: string) => contactId && chat.react(contactId, messageId, emoji),
    setTyping: (isTypingNow: boolean) => contactId && chat.setTyping(contactId, isTypingNow),
    fetchMore: (page: number) => contactId && chat.fetchHistory(contactId, page),
  };
}
