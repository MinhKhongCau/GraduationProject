"use client";

import { useEffect, useState } from "react";
import { ContactList } from "./component/ContactList";
import { ChatWindow } from "./component/ChatWindow";
import { useExpertChatContacts, useDmThread, useKeyboardHeight } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";
import { useChatContext } from "@/context/ChatContext";

export default function ExpertMessagesPage() {
  const { user } = useAuthContext();
  const { data: contacts = [] } = useExpertChatContacts();
  const { onlineByContactId, queryPresence } = useChatContext();
  const [activeContactId, setActiveContactId] = useState<string | null>(null);
  const keyboardHeight = useKeyboardHeight();

  useEffect(() => {
    if (!activeContactId && contacts.length > 0) setActiveContactId(contacts[0].id);
  }, [contacts, activeContactId]);

  useEffect(() => {
    if (contacts.length > 0) queryPresence(contacts.map((c) => c.id));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [contacts]);

  const thread = useDmThread(activeContactId);
  const activeContact = contacts.find((c) => c.id === activeContactId) ?? null;

  return (
    <div
      className="flex h-[calc(100dvh-9.5rem-env(safe-area-inset-top)-env(safe-area-inset-bottom)-var(--keyboard-height))] overflow-hidden rounded-2xl border border-border bg-background shadow-card sm:h-[750px]"
      style={{ "--keyboard-height": `${keyboardHeight}px` } as React.CSSProperties}
    >
      <ContactList
        contacts={contacts}
        activeContactId={activeContactId}
        onlineByContactId={onlineByContactId}
        onSelect={(contact) => setActiveContactId(contact.id)}
      />
      {activeContact && user ? (
        <ChatWindow
          contact={activeContact}
          messages={thread.messages}
          myId={user.id}
          isTyping={thread.isTyping}
          online={thread.online}
          reactionsByMessageId={thread.reactionsByMessageId}
          onSend={thread.sendText}
          onSendVoice={thread.sendVoice}
          onReact={thread.react}
          onTyping={thread.setTyping}
        />
      ) : (
        <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
          {contacts.length === 0 ? "No conversations yet." : "Select a conversation"}
        </div>
      )}
    </div>
  );
}
