"use client";

import { useState } from "react";
import { Button, Textarea } from "@/components/ui";
import { useCreateComment } from "@/hooks";

export interface CommentComposerProps {
  postId: number;
  parentId?: number | null;
  placeholder?: string;
  onDone?: () => void;
}

export function CommentComposer({ postId, parentId = null, placeholder = "Write a comment...", onDone }: CommentComposerProps) {
  const [content, setContent] = useState("");
  const createComment = useCreateComment(postId);

  const submit = () => {
    const trimmed = content.trim();
    if (!trimmed) return;
    createComment.mutate(
      { content: trimmed, parentId },
      {
        onSuccess: () => {
          setContent("");
          onDone?.();
        },
      }
    );
  };

  return (
    <div className="flex flex-col gap-2">
      <Textarea
        aria-label={placeholder}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        rows={2}
        className="resize-none"
      />
      <div className="flex justify-end gap-2">
        {onDone && (
          <Button variant="ghost" size="sm" onClick={onDone}>
            Cancel
          </Button>
        )}
        <Button size="sm" onClick={submit} disabled={createComment.isPending || !content.trim()}>
          {createComment.isPending ? "Posting..." : "Post"}
        </Button>
      </div>
    </div>
  );
}
