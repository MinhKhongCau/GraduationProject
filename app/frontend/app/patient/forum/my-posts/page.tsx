"use client";

import Link from "next/link";
import { Edit2, Ban } from "lucide-react";
import { Card, Badge, Button, PageHeader, Spinner, buttonClasses, type BadgeTone } from "@/components/ui";
import { useMyPosts, useChangePostStatus } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";
import { ROUTES } from "@/constants";
import type { Post, PostStatus } from "@/types";

const STATUS_TONE: Record<PostStatus, BadgeTone> = {
  PUBLISHED: "success",
  DRAFT: "neutral",
  ARCHIVED: "danger",
};

const STATUS_LABEL: Record<PostStatus, string> = {
  PUBLISHED: "Published",
  DRAFT: "Draft",
  ARCHIVED: "Blocked",
};

export default function MyPostsPage() {
  const { user } = useAuthContext();
  const { data: posts = [], isLoading } = useMyPosts(user?.id);
  const changeStatus = useChangePostStatus();

  const handleBlock = (post: Post) => {
    if (confirm(`Block "${post.title}"? This hides it from the community and can't be undone.`)) {
      changeStatus.mutate({ postId: post.id, status: "ARCHIVED" });
    }
  };

  return (
    <div className="mx-auto max-w-4xl">
      <PageHeader
        title="My posts"
        actions={
          <Link href={ROUTES.PATIENT.FORUM_NEW_POST} className={buttonClasses("primary")}>
            New post
          </Link>
        }
      />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : posts.length === 0 ? (
        <Card className="p-10 text-center text-sm text-muted-foreground">You haven&apos;t written any posts yet.</Card>
      ) : (
        <div className="space-y-3">
          {posts.map((post) => (
            <Card
              key={post.id}
              className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0 flex-1">
                <div className="mb-1 flex flex-wrap items-center gap-2">
                  <Link
                    href={ROUTES.PATIENT.FORUM_POST(post.slug)}
                    className="font-semibold text-foreground transition-colors hover:text-primary"
                  >
                    {post.title}
                  </Link>
                  <Badge tone={STATUS_TONE[post.status]}>{STATUS_LABEL[post.status]}</Badge>
                </div>
                {post.summary && <p className="line-clamp-1 text-sm text-muted-foreground">{post.summary}</p>}
              </div>
              <div className="flex flex-shrink-0 items-center gap-2">
                <Link href={ROUTES.PATIENT.FORUM_EDIT_POST(post.slug)} className={buttonClasses("outline", "sm")}>
                  <Edit2 className="h-3.5 w-3.5" aria-hidden="true" /> Edit
                </Link>
                {post.status !== "ARCHIVED" && (
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={changeStatus.isPending}
                    onClick={() => handleBlock(post)}
                    className="text-danger hover:bg-danger-soft"
                  >
                    <Ban className="h-3.5 w-3.5" /> Block
                  </Button>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
