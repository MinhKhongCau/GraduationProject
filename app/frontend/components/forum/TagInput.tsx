"use client";

import { useState, type KeyboardEvent } from "react";
import { X } from "lucide-react";

export interface TagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
}

/** Hashtag-style tag entry: typing "#<tag>" then space/enter/comma commits it as a chip. */
export function TagInput({ value, onChange, placeholder }: TagInputProps) {
  const [draft, setDraft] = useState("");

  const commitDraft = () => {
    const raw = draft.trim();
    setDraft("");
    if (!raw.startsWith("#")) return;
    const tag = raw.replace(/^#/, "");
    if (!tag || value.includes(tag)) return;
    onChange([...value, tag]);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter" || event.key === " " || event.key === ",") {
      if (draft.trim().startsWith("#")) {
        event.preventDefault();
        commitDraft();
      } else if (event.key === "Enter") {
        event.preventDefault();
      }
      return;
    }
    if (event.key === "Backspace" && draft === "" && value.length > 0) {
      onChange(value.slice(0, -1));
    }
  };

  const removeTag = (tag: string) => onChange(value.filter((t) => t !== tag));

  return (
    <div className="flex min-h-10 flex-wrap items-center gap-1.5 rounded-lg border border-border-strong bg-background px-3 py-1.5 transition-colors focus-within:border-primary focus-within:ring-2 focus-within:ring-primary/20">
      {value.map((tag) => (
        <span
          key={tag}
          className="flex items-center gap-1 rounded-full bg-primary-soft px-2 py-0.5 text-xs font-medium text-primary-soft-text"
        >
          #{tag}
          <button
            type="button"
            onClick={() => removeTag(tag)}
            aria-label={`Remove tag ${tag}`}
            className="inline-flex h-4 w-4 items-center justify-center rounded-full text-primary-soft-text/70 hover:bg-primary/10 hover:text-primary-soft-text"
          >
            <X className="h-3 w-3" />
          </button>
        </span>
      ))}
      <input
        type="text"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={commitDraft}
        placeholder={value.length === 0 ? placeholder : undefined}
        aria-label={placeholder ?? "Add tag"}
        className="min-w-[120px] flex-1 border-none bg-transparent py-0.5 text-sm text-foreground outline-none"
      />
    </div>
  );
}
