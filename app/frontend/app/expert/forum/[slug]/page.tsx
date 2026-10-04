"use client";

import { useParams } from "next/navigation";
import { PostContent, CommentThread, CommentComposer, LikeButton, BookmarkButton } from "@/components/forum";
import { Card, Spinner } from "@/components/ui";
import { useForumPost, useForumComments } from "@/hooks";

export default function ExpertForumPostPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: post, isLoading } = useForumPost(slug);
  const { data: comments = [] } = useForumComments(post?.id ?? 0);

  if (isLoading) return <Spinner className="h-6 w-6" />;
  if (!post) return <p className="text-sm text-muted-foreground">Post not found.</p>;

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Card className="p-6">
        <PostContent post={post} />
        <div className="mt-6 flex gap-2">
          <LikeButton postId={post.id} initialLikeCount={post.likeCount} />
          <BookmarkButton postId={post.id} initialBookmarkCount={post.bookmarkCount} />
        </div>
      </Card>

      <Card className="p-6">
        <h2 className="mb-4 text-lg font-semibold text-foreground">Comments ({post.commentCount})</h2>
        <div className="mb-6">
          <CommentComposer postId={post.id} />
        </div>
        <CommentThread postId={post.id} comments={comments} />
      </Card>
    </div>
  );
}
