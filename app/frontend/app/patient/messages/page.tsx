"use client";

import { useEffect, useState } from "react";
import { ContactList } from "./component/ContactList";
import { ChatWindow } from "./component/ChatWindow";
import { usePatientChatContacts, useDmThread, useKeyboardHeight } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";
import { useChatContext } from "@/context/ChatContext";
import { ChatContact } from "@/types";
import { MessageCircle } from "lucide-react";

export default function MessagesPage() {
  const { user } = useAuthContext();
  const { data: contacts = [] } = usePatientChatContacts();
  const { onlineByContactId, queryPresence } = useChatContext();
  const [activeContactId, setActiveContactId] = useState<string | null>(null);
  const [showChatOnMobile, setShowChatOnMobile] = useState(false);
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

  const handleSelectContact = (contact: ChatContact) => {
    setActiveContactId(contact.id);
    setShowChatOnMobile(true);
  };

  return (
    <div
      className="-mx-4 -mt-5 -mb-[calc(2.5rem-1px)] flex h-[calc(100dvh-7.0625rem-env(safe-area-inset-top)-env(safe-area-inset-bottom)-var(--keyboard-height))] overflow-hidden bg-background sm:mx-0 sm:mt-0 sm:mb-0 sm:h-[calc(100dvh-7.5rem)] sm:min-h-[520px] sm:rounded-2xl sm:border sm:border-border/70 sm:shadow-elevated lg:h-[calc(100dvh-8.5rem)]"
      style={{ "--keyboard-height": `${keyboardHeight}px` } as React.CSSProperties}
    >
      <ContactList
        contacts={contacts}
        activeContactId={activeContactId}
        onlineByContactId={onlineByContactId}
        onSelect={handleSelectContact}
        className={showChatOnMobile ? "hidden md:flex" : "flex"}
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
          className={showChatOnMobile ? "flex" : "hidden md:flex"}
          onBack={() => setShowChatOnMobile(false)}
          onFetchMore={thread.fetchMore}
        />
      ) : (
        <div className={`flex-1 flex-col items-center justify-center gap-3 bg-gradient-to-b from-surface/50 to-background p-8 text-center ${showChatOnMobile ? "flex" : "hidden md:flex"}`}>
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-primary-soft text-primary">
            <MessageCircle className="h-7 w-7" aria-hidden="true" />
          </span>
          <p className="text-sm font-semibold text-foreground">
            {contacts.length === 0 ? "No conversations yet" : "Select a conversation"}
          </p>
          <p className="max-w-xs text-xs text-muted-foreground">
            {contacts.length === 0 ? "Book an expert to start chatting." : "Choose a contact on the left to continue chatting."}
          </p>
        </div>
      )}
    </div>
  );
}
