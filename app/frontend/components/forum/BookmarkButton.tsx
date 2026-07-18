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
      onClick={onClick}
      disabled={toggleBookmark.isPending}
      className={`flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
        bookmarked ? "border-primary bg-primary-soft text-primary" : "border-border text-muted-foreground hover:bg-surface"
      }`}
    >
      <Bookmark className={`h-4 w-4 ${bookmarked ? "fill-primary" : ""}`} />
      {count}
    </button>
  );
}
