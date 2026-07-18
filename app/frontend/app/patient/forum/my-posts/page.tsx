"use client";

import Link from "next/link";
import { Edit2, Ban } from "lucide-react";
import { Card, Button, Spinner } from "@/components/ui";
import { useMyPosts, useChangePostStatus } from "@/hooks";
import { useAuthContext } from "@/context/AuthContext";
import { ROUTES } from "@/constants";
import type { Post, PostStatus } from "@/types";

const STATUS_BADGE: Record<PostStatus, string> = {
  PUBLISHED: "bg-success-soft text-success",
  DRAFT: "bg-surface text-muted-foreground",
  ARCHIVED: "bg-danger-soft text-danger",
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
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">My posts</h1>
        <Link href={ROUTES.PATIENT.FORUM_NEW_POST}>
          <Button size="sm">New post</Button>
        </Link>
      </div>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : posts.length === 0 ? (
        <p className="text-sm text-muted-foreground">You haven&apos;t written any posts yet.</p>
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
                    className="font-bold text-foreground hover:text-primary"
                  >
                    {post.title}
                  </Link>
                  <span className={`rounded-full px-2 py-0.5 text-[11px] font-medium ${STATUS_BADGE[post.status]}`}>
                    {STATUS_LABEL[post.status]}
                  </span>
                </div>
                {post.summary && <p className="line-clamp-1 text-sm text-muted-foreground">{post.summary}</p>}
              </div>
              <div className="flex flex-shrink-0 items-center gap-2">
                <Link href={ROUTES.PATIENT.FORUM_EDIT_POST(post.slug)}>
                  <Button variant="outline" size="sm" className="flex items-center gap-1">
                    <Edit2 className="h-3.5 w-3.5" /> Edit
                  </Button>
                </Link>
                {post.status !== "ARCHIVED" && (
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={changeStatus.isPending}
                    onClick={() => handleBlock(post)}
                    className="flex items-center gap-1 text-danger hover:bg-danger-soft/50"
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
