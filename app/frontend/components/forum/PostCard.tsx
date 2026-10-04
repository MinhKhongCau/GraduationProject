import Link from "next/link";
import { Heart, MessageSquare, Bookmark } from "lucide-react";
import { Badge, Card } from "@/components/ui";
import type { Post } from "@/types";
import { useForumAuthorName } from "@/hooks";

export interface PostCardProps {
  post: Post;
  /** "/patient/forum" or "/expert/forum" — post detail routes are role-scoped. */
  basePath: string;
}

function initials(fullName: string): string {
  return fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function PostCard({ post, basePath }: PostCardProps) {
  const fallbackAuthorName = useForumAuthorName(post.authorId);
  const authorName = post.author?.name || fallbackAuthorName;
  const avatarUrl = post.author?.avatarUrl;

  return (
    <Card className="flex flex-col overflow-hidden p-0 transition-shadow hover:shadow-elevated">
      {post.thumbnailUrl && (
        <Link href={`${basePath}/${post.slug}`} className="block aspect-[16/9] w-full overflow-hidden border-b border-border">
          <img
            src={post.thumbnailUrl}
            alt={post.title}
            className="h-full w-full object-cover transition-transform duration-300 hover:scale-105"
          />
        </Link>
      )}

      <div className="flex flex-1 flex-col p-5">
        {/* Author info row */}
        <div className="mb-4 flex items-center gap-3">
          <div className="relative flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-surface text-xs font-bold text-muted-foreground">
            {avatarUrl ? (
              <img src={avatarUrl} alt={authorName} className="h-full w-full object-cover" />
            ) : (
              initials(authorName) || "?"
            )}
          </div>
          <div className="flex flex-col min-w-0">
            <span className="truncate text-sm font-semibold text-foreground">{authorName}</span>
            <span className="text-xs text-muted-foreground">
              {new Date(post.createdAt).toLocaleDateString("vi-VN")}
            </span>
          </div>
        </div>

        <Link href={`${basePath}/${post.slug}`} className="mb-2">
          <h3 className="text-base font-semibold leading-snug text-foreground hover:text-primary line-clamp-2">{post.title}</h3>
        </Link>

        {post.summary && <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">{post.summary}</p>}

        {post.tags.length > 0 && (
          <div className="mb-4 flex flex-wrap gap-1.5">
            {post.tags.map((tag) => (
              <Badge key={tag} tone="primary">
                {tag}
              </Badge>
            ))}
          </div>
        )}

        <div className="mt-auto flex items-center gap-4 border-t border-border pt-3 text-xs font-medium text-muted-foreground">
          <span className="flex items-center gap-1">
            <Heart className="h-3.5 w-3.5" /> {post.likeCount}
          </span>
          <span className="flex items-center gap-1">
            <MessageSquare className="h-3.5 w-3.5" /> {post.commentCount}
          </span>
          <span className="flex items-center gap-1">
            <Bookmark className="h-3.5 w-3.5" /> {post.bookmarkCount}
          </span>
        </div>
      </div>
    </Card>
  );
}
