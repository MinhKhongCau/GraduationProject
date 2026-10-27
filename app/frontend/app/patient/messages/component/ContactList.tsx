"use client";

import Image from "next/image";
import { MessageCircle, Search } from "lucide-react";
import type { ChatContact } from "@/types";
import { resolveImageUrl } from "@/lib/image";

export interface ContactListProps {
  contacts: ChatContact[];
  activeContactId: string | null;
  onlineByContactId: Record<string, boolean>;
  onSelect: (contact: ChatContact) => void;
  className?: string;
}

function roleLabel(role: ChatContact["role"]): string {
  if (role === "EXPERT") return "Expert";
  if (role === "CHATBOT") return "AI Assistant";
  return "Patient";
}

export function ContactList({ contacts, activeContactId, onlineByContactId, onSelect, className }: ContactListProps) {
  return (
    <div className={`w-full shrink-0 flex-col bg-background md:w-80 md:border-r md:border-border/70 ${className ?? ""}`}>
      <div className="px-4 pt-4 pb-3">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-xl font-bold tracking-tight text-foreground">Messages</h2>
          {contacts.length > 0 && (
            <span className="rounded-full bg-primary-soft px-2.5 py-0.5 text-xs font-semibold text-primary-soft-text">{contacts.length}</span>
          )}
        </div>
        <div className="relative">
          <Search className="pointer-events-none absolute top-1/2 left-3.5 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <input
            type="text"
            aria-label="Search experts"
            placeholder="Search experts..."
            className="h-10 w-full rounded-full border border-transparent bg-surface pr-4 pl-10 text-sm text-foreground transition-colors placeholder:text-muted-foreground/70 focus:border-primary/40 focus:bg-background focus:ring-2 focus:ring-primary/15 focus:outline-none"
          />
        </div>
      </div>

      <div className="flex-1 space-y-0.5 overflow-y-auto px-2 pb-3">
        {contacts.length === 0 && (
          <div className="flex flex-col items-center gap-2 px-4 py-10 text-center">
            <span className="flex h-12 w-12 items-center justify-center rounded-full bg-surface text-muted-foreground">
              <MessageCircle className="h-5 w-5" aria-hidden="true" />
            </span>
            <p className="text-sm text-muted-foreground">No conversations yet.</p>
          </div>
        )}
        {contacts.map((contact) => {
          const isActive = activeContactId === contact.id;
          const online = !!onlineByContactId[contact.id] || contact.role === "CHATBOT";
          return (
            <button
              key={contact.id}
              type="button"
              onClick={() => onSelect(contact)}
              aria-current={isActive ? "true" : undefined}
              className={`flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-colors ${
                isActive ? "bg-primary-soft" : "hover:bg-surface active:bg-surface"
              }`}
            >
              <div className="relative shrink-0">
                <Image
                  src={resolveImageUrl(contact.avatarUrl) || "https://i.pravatar.cc/150?u=" + contact.id}
                  alt={contact.fullName}
                  width={48}
                  height={48}
                  unoptimized
                  className={`h-12 w-12 rounded-full object-cover ring-2 ${isActive ? "ring-primary/30" : "ring-background"}`}
                />
                {online && <span className="absolute right-0 bottom-0 h-3.5 w-3.5 rounded-full border-[2.5px] border-background bg-success" />}
              </div>
              <div className="min-w-0 flex-1">
                <h3 className={`truncate text-sm font-semibold ${isActive ? "text-primary-soft-text" : "text-foreground"}`}>{contact.fullName}</h3>
                <p className="mt-0.5 flex items-center gap-1.5 truncate text-xs text-muted-foreground">
                  <span>{roleLabel(contact.role)}</span>
                  {online && (
                    <>
                      <span aria-hidden="true">·</span>
                      <span className="text-success">Online</span>
                    </>
                  )}
                </p>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}
