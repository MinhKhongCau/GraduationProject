"use client";

import Image from "next/image";
import { Search } from "lucide-react";
import type { MockContact } from "@/data";

export interface ContactListProps {
  contacts: MockContact[];
  activeContactId: string;
  onSelect: (contact: MockContact) => void;
}

export function ContactList({ contacts, activeContactId, onSelect }: ContactListProps) {
  return (
    <div className="flex w-80 shrink-0 flex-col border-r border-border bg-background">
      <div className="border-b border-border p-4">
        <h2 className="mb-4 text-xl font-bold text-foreground">Messages</h2>
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
            <Search className="h-4 w-4" />
          </div>
          <input
            type="text"
            className="block w-full rounded-xl border-none bg-surface py-2 pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-primary-soft"
            placeholder="Search experts..."
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {contacts.map((contact) => (
          <button
            key={contact.id}
            onClick={() => onSelect(contact)}
            className={`flex w-full items-center gap-3 border-b border-border/50 p-4 text-left transition-colors ${
              activeContactId === contact.id ? "border-r-2 border-r-primary bg-primary-soft/50" : "hover:bg-surface"
            }`}
          >
            <div className="relative shrink-0">
              <Image
                src={contact.avatar}
                alt={contact.name}
                width={48}
                height={48}
                unoptimized
                className="h-12 w-12 rounded-full border border-border object-cover"
              />
              {contact.online && (
                <span className="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-background bg-success" />
              )}
            </div>
            <div className="min-w-0 flex-1">
              <div className="mb-0.5 flex items-baseline justify-between">
                <h3
                  className={`truncate text-sm ${contact.unread ? "font-bold text-foreground" : "font-semibold text-foreground/80"}`}
                >
                  {contact.name}
                </h3>
                <span className="ml-2 shrink-0 text-xs text-muted-foreground">{contact.time}</span>
              </div>
              <div className="flex items-center justify-between">
                <p className={`truncate text-xs ${contact.unread ? "font-semibold text-foreground/80" : "text-muted-foreground"}`}>
                  {contact.lastMessage}
                </p>
                {contact.unread && <span className="ml-2 h-2 w-2 shrink-0 rounded-full bg-primary" />}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
