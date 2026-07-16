"use client";

import { useEffect, useState } from "react";
import { ContactList } from "./component/ContactList";
import { ChatWindow } from "./component/ChatWindow";
import { usePatientChatContacts, useDmThread } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";
import { useChatContext } from "@/context/ChatContext";

export default function MessagesPage() {
  const { user } = useAuthContext();
  const { data: contacts = [] } = usePatientChatContacts();
  const { onlineByContactId, queryPresence } = useChatContext();
  const [activeContactId, setActiveContactId] = useState<string | null>(null);

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
    <div className="flex h-[750px] overflow-hidden rounded-2xl border border-border bg-background shadow-card">
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
          {contacts.length === 0 ? "No conversations yet — book an expert to start chatting." : "Select a conversation"}
        </div>
      )}
    </div>
  );
}
