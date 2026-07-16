import { useForumAuthorName } from "@/hooks";
import type { PostDetail } from "@/types";

export interface PostContentProps {
  post: PostDetail;
}

export function PostContent({ post }: PostContentProps) {
  const authorName = useForumAuthorName(post.authorId);

  return (
    <div>
      <h1 className="mb-2 text-2xl font-bold text-foreground">{post.title}</h1>
      <div className="mb-4 flex items-center gap-2 text-xs text-muted-foreground">
        <span>{authorName}</span>
        <span>•</span>
        <span>{new Date(post.createdAt).toLocaleDateString()}</span>
        <span>•</span>
        <span>{post.viewCount} views</span>
      </div>
      {post.tags.length > 0 && (
        <div className="mb-6 flex flex-wrap gap-1.5">
          {post.tags.map((tag) => (
            <span key={tag} className="rounded-full bg-primary-soft px-2 py-0.5 text-[11px] font-medium text-primary-soft-text">
              {tag}
            </span>
          ))}
        </div>
      )}
      {/* forum-service's post content is plain text (no markdown renderer
       * wired up yet) — whitespace-pre-wrap keeps paragraph breaks readable. */}
      <div className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">{post.content}</div>
    </div>
  );
}
