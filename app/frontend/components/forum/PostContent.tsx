"use client";

import dynamic from "next/dynamic";
import { useForumAuthorName } from "@/hooks";
import type { PostDetail } from "@/types";
import { Badge } from "@/components/ui";

const ContentViewer = dynamic(() => import("@/components/ui/MDXEditor"), { ssr: false });

export interface PostContentProps {
  post: PostDetail;
}

function initials(fullName: string): string {
  return fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function PostContent({ post }: PostContentProps) {
  const fallbackAuthorName = useForumAuthorName(post.authorId);
  const authorName = post.author?.name || fallbackAuthorName;
  const avatarUrl = post.author?.avatarUrl;

  return (
    <div>
      <h1 className="mb-4 text-2xl font-bold leading-tight tracking-tight text-foreground sm:text-3xl">{post.title}</h1>
      
      {/* Author and metadata */}
      <div className="mb-6 flex items-center gap-3 border-b border-border pb-4">
        <div className="relative flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-surface text-sm font-bold text-muted-foreground">
          {avatarUrl ? (
            <img src={avatarUrl} alt={authorName} className="h-full w-full object-cover" />
          ) : (
            initials(authorName) || "?"
          )}
        </div>
        <div className="flex flex-col min-w-0">
          <span className="text-sm font-semibold text-foreground">{authorName}</span>
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span>{new Date(post.createdAt).toLocaleDateString("vi-VN")}</span>
            <span aria-hidden="true">•</span>
            <span>{post.viewCount} lượt xem</span>
          </div>
        </div>
      </div>

      {post.tags.length > 0 && (
        <div className="mb-6 flex flex-wrap gap-1.5">
          {post.tags.map((tag) => (
            <Badge key={tag} tone="primary">
              {tag}
            </Badge>
          ))}
        </div>
      )}

      {post.thumbnailUrl && (
        <div className="mb-6 aspect-[16/9] w-full overflow-hidden rounded-xl border border-border">
          <img
            src={post.thumbnailUrl}
            alt={post.title}
            className="h-full w-full object-cover"
          />
        </div>
      )}

      <ContentViewer value={post.content} readOnly />
    </div>
  );
}
