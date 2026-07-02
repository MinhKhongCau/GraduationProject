"use client";

import { useState } from "react";
import { ContactList } from "./component/ContactList";
import { ChatWindow } from "./component/ChatWindow";
import { CONTACTS_MOCK, CHAT_HISTORY_MOCK } from "@/data";

/** No chat backend exists yet (see DESIGN.md) — UI-only with mock data. */
export default function MessagesPage() {
  const [activeContact, setActiveContact] = useState(CONTACTS_MOCK[0]);

  return (
    <div className="flex h-[750px] overflow-hidden rounded-2xl border border-border bg-background shadow-card">
      <ContactList contacts={CONTACTS_MOCK} activeContactId={activeContact.id} onSelect={setActiveContact} />
      <ChatWindow contact={activeContact} messages={CHAT_HISTORY_MOCK} />
    </div>
  );
}
