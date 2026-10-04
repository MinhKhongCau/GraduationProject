"use client";

import { useState } from "react";
import { Bookmark } from "lucide-react";
import { useToggleBookmark } from "@/hooks";

export interface BookmarkButtonProps {
  postId: number;
  initialBookmarkCount: number;
}

/** Same gap as LikeButton: no GET endpoint exposes whether the current user
 * already bookmarked this specific post, so this starts assuming "not
 * bookmarked" every page load. */
export function BookmarkButton({ postId, initialBookmarkCount }: BookmarkButtonProps) {
  const [bookmarked, setBookmarked] = useState(false);
  const [count, setCount] = useState(initialBookmarkCount);
  const toggleBookmark = useToggleBookmark(postId);

  const onClick = () => {
    toggleBookmark.mutate(bookmarked, {
      onSuccess: (result) => {
        setBookmarked(result.bookmarked);
        setCount(result.bookmarkCount);
      },
    });
  };

  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={bookmarked}
      aria-label="Bookmark"
      disabled={toggleBookmark.isPending}
      className={`inline-flex h-9 items-center gap-1.5 rounded-lg border px-3 text-sm font-medium transition-colors disabled:opacity-60 ${
        bookmarked ? "border-primary/30 bg-primary-soft text-primary" : "border-border-strong bg-background text-muted-foreground hover:bg-surface hover:text-foreground"
      }`}
    >
      <Bookmark aria-hidden="true" className={`h-4 w-4 ${bookmarked ? "fill-primary" : ""}`} />
      {count}
    </button>
  );
}
