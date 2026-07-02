import Image from "next/image";
import { Phone, Video, MoreVertical, Paperclip, SendHorizonal, Circle } from "lucide-react";
import { MessageBubble } from "./MessageBubble";
import type { MockContact, MockChatMessage } from "@/data";

export interface ChatWindowProps {
  contact: MockContact;
  messages: MockChatMessage[];
}

export function ChatWindow({ contact, messages }: ChatWindowProps) {
  return (
    <div className="flex min-w-0 flex-1 flex-col bg-background">
      <div className="flex items-center justify-between border-b border-border p-4">
        <div className="flex items-center gap-3">
          <div className="relative">
            <Image
              src={contact.avatar}
              alt={contact.name}
              width={40}
              height={40}
              unoptimized
              className="h-10 w-10 rounded-full object-cover"
            />
            {contact.online && (
              <span className="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border-2 border-background bg-success" />
            )}
          </div>
          <div>
            <h3 className="text-sm font-bold text-foreground">{contact.name}</h3>
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <span>Clinical Psychologist</span>
              {contact.online && (
                <>
                  <Circle className="h-1 w-1 fill-border text-border" />
                  <span className="font-medium text-primary">Active now</span>
                </>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2 text-muted-foreground">
          <button className="rounded-full p-2 transition-colors hover:bg-surface">
            <Phone className="h-5 w-5" />
          </button>
          <button className="rounded-full p-2 transition-colors hover:bg-surface">
            <Video className="h-5 w-5" />
          </button>
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
          <MessageBubble key={message.id} message={message} />
        ))}
      </div>

      <div className="border-t border-border bg-background p-4">
        <div className="flex items-center gap-2 rounded-2xl border border-border bg-surface p-2">
          <button className="shrink-0 rounded-full p-2 text-muted-foreground transition-colors hover:bg-border/50">
            <Paperclip className="h-5 w-5" />
          </button>
          <input
            type="text"
            placeholder="Type your message..."
            className="flex-1 border-none bg-transparent px-2 text-sm outline-none"
          />
          <button className="shrink-0 rounded-full p-2 text-muted-foreground transition-colors hover:bg-primary-soft hover:text-primary">
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
