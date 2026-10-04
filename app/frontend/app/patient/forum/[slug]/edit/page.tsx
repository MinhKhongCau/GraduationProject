"use client";

import { useParams } from "next/navigation";
import { PostForm } from "@/components/forum";
import { Card, PageHeader, Spinner } from "@/components/ui";
import { ROUTES } from "@/constants";
import { useForumPost } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";

export default function EditForumPostPage() {
  const { slug } = useParams<{ slug: string }>();
  const { user } = useAuthContext();
  const { data: post, isLoading } = useForumPost(slug);

  if (isLoading) return <Spinner className="h-6 w-6" />;
  if (!post) return <p className="text-sm text-muted-foreground">Post not found.</p>;
  if (user && post.authorId !== user.id) {
    return <p className="text-sm text-muted-foreground">You can only edit your own posts.</p>;
  }

  return (
    <div className="mx-auto max-w-2xl">
      <PageHeader title="Edit post" />
      <Card className="p-5 sm:p-6">
        <PostForm
          basePath={ROUTES.PATIENT.FORUM}
          mode="edit"
          postId={post.id}
          initialValues={{
            categoryId: post.categoryId,
            title: post.title,
            summary: post.summary,
            content: post.content,
            tags: post.tags,
          }}
        />
      </Card>
    </div>
  );
}
