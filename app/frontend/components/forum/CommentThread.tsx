"use client";

import { useState } from "react";
import { useForumAuthor } from "@/hooks";
import { CommentComposer } from "./CommentComposer";
import type { Comment } from "@/types";

export interface CommentThreadProps {
  postId: number;
  comments: Comment[];
}

function initials(fullName: string): string {
  return fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function CommentThread({ postId, comments }: CommentThreadProps) {
  if (comments.length === 0) {
    return <p className="text-sm text-muted-foreground">Chưa có bình luận nào — hãy là người đầu tiên trả lời.</p>;
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
  const clientProfile = useForumAuthor(comment.user ? "" : comment.userId);
  const authorName = comment.user?.name ?? clientProfile.name;
  const avatarUrl = comment.user?.avatarUrl ?? clientProfile.avatarUrl;

  return (
    <div>
      <div className="rounded-xl border border-border bg-surface/50 p-3">
        <div className="mb-2 flex items-center gap-2">
          <div className="relative flex h-6 w-6 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-surface text-[10px] font-bold text-muted-foreground">
            {avatarUrl ? (
              <img src={avatarUrl} alt={authorName} className="h-full w-full object-cover" />
            ) : (
              initials(authorName) || "?"
            )}
          </div>
          <div className="flex items-baseline gap-1.5 min-w-0">
            <span className="truncate text-xs font-bold text-foreground">{authorName}</span>
            <span className="text-[10px] text-muted-foreground">
              {new Date(comment.createdAt).toLocaleString("vi-VN")}
            </span>
          </div>
        </div>
        <p className={`text-sm pl-8 ${comment.deleted ? "italic text-muted-foreground" : "text-foreground"}`}>
          {comment.deleted ? "[bình luận đã bị xóa]" : comment.content}
        </p>
        {!comment.deleted && (
          <button onClick={() => setReplying((v) => !v)} className="mt-1 ml-8 text-xs font-medium text-primary hover:underline">
            Phản hồi
          </button>
        )}
      </div>

      {replying && (
        <div className="ml-8 mt-2">
          <CommentComposer postId={postId} parentId={comment.id} placeholder="Viết phản hồi..." onDone={() => setReplying(false)} />
        </div>
      )}

      {comment.replies.length > 0 && (
        <div className="ml-8 mt-3 space-y-3 border-l border-border pl-3">
          {comment.replies.map((reply) => (
            <CommentNode key={reply.id} postId={postId} comment={reply} />
          ))}
        </div>
      )}
    </div>
  );
}
