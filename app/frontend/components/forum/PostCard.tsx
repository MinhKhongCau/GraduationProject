import Link from "next/link";
import { Heart, MessageSquare, Bookmark } from "lucide-react";
import { Card } from "@/components/ui";
import type { Post } from "@/types";

export interface PostCardProps {
  post: Post;
  /** "/patient/forum" or "/expert/forum" — post detail routes are role-scoped. */
  basePath: string;
}

export function PostCard({ post, basePath }: PostCardProps) {
  return (
    <Card className="flex flex-col p-6">
      <Link href={`${basePath}/${post.slug}`} className="mb-2">
        <h3 className="font-bold text-foreground hover:text-primary">{post.title}</h3>
      </Link>
      {post.summary && <p className="mb-3 line-clamp-2 text-sm text-muted-foreground">{post.summary}</p>}
      {post.tags.length > 0 && (
        <div className="mb-3 flex flex-wrap gap-1.5">
          {post.tags.map((tag) => (
            <span key={tag} className="rounded-full bg-primary-soft px-2 py-0.5 text-[11px] font-medium text-primary-soft-text">
              {tag}
            </span>
          ))}
        </div>
      )}
      <div className="mt-auto flex items-center gap-4 text-xs text-muted-foreground">
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
    </Card>
  );
}
