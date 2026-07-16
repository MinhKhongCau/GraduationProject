"use client";

import { useState } from "react";
import { useForumAuthorName } from "@/hooks";
import { CommentComposer } from "./CommentComposer";
import type { Comment } from "@/types";

export interface CommentThreadProps {
  postId: number;
  comments: Comment[];
}

export function CommentThread({ postId, comments }: CommentThreadProps) {
  if (comments.length === 0) {
    return <p className="text-sm text-muted-foreground">No comments yet — be the first to reply.</p>;
  }
  return (
    <div className="space-y-4">
      {comments.map((comment) => (
        <CommentNode key={comment.id} postId={postId} comment={comment} />
      ))}
    </div>
  );
}

function CommentNode({ postId, comment }: { postId: number; comment: Comment }) {
  const [replying, setReplying] = useState(false);
  const authorName = useForumAuthorName(comment.userId);

  return (
    <div>
      <div className="rounded-xl border border-border bg-surface/50 p-3">
        <div className="mb-1 flex items-center gap-2 text-xs">
          <span className="font-semibold text-foreground">{authorName}</span>
          <span className="text-muted-foreground">{new Date(comment.createdAt).toLocaleString()}</span>
        </div>
        <p className={`text-sm ${comment.deleted ? "italic text-muted-foreground" : "text-foreground"}`}>
          {comment.deleted ? "[comment deleted]" : comment.content}
        </p>
        {!comment.deleted && (
          <button onClick={() => setReplying((v) => !v)} className="mt-1 text-xs font-medium text-primary hover:underline">
            Reply
          </button>
        )}
      </div>

      {replying && (
        <div className="ml-6 mt-2">
          <CommentComposer postId={postId} parentId={comment.id} placeholder="Write a reply..." onDone={() => setReplying(false)} />
        </div>
      )}

      {comment.replies.length > 0 && (
        <div className="ml-6 mt-3 space-y-3 border-l border-border pl-3">
          {comment.replies.map((reply) => (
            <CommentNode key={reply.id} postId={postId} comment={reply} />
          ))}
        </div>
      )}
    </div>
  );
}
