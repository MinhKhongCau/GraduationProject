"use client";

import { useState } from "react";
import { Heart } from "lucide-react";
import { useToggleLike } from "@/hooks";

export interface LikeButtonProps {
  postId: number;
  initialLikeCount: number;
}

/** GAP: forum-service's GET endpoints don't expose whether the current user
 * already liked a post (only the POST/DELETE like actions return `liked`),
 * so this starts assuming "not liked" every page load rather than
 * reflecting real prior-session state. */
export function LikeButton({ postId, initialLikeCount }: LikeButtonProps) {
  const [liked, setLiked] = useState(false);
  const [count, setCount] = useState(initialLikeCount);
  const toggleLike = useToggleLike(postId);

  const onClick = () => {
    toggleLike.mutate(liked, {
      onSuccess: (result) => {
        setLiked(result.liked);
        setCount(result.likeCount);
      },
    });
  };

  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={liked}
      aria-label="Like"
      disabled={toggleLike.isPending}
      className={`inline-flex h-9 items-center gap-1.5 rounded-lg border px-3 text-sm font-medium transition-colors disabled:opacity-60 ${
        liked ? "border-primary/30 bg-primary-soft text-primary" : "border-border-strong bg-background text-muted-foreground hover:bg-surface hover:text-foreground"
      }`}
    >
      <Heart aria-hidden="true" className={`h-4 w-4 ${liked ? "fill-primary" : ""}`} />
      {count}
    </button>
  );
}
