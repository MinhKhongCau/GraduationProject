"use client";

import { useState } from "react";
import { PostCard } from "@/components/forum";
import { Card, PageHeader, Spinner, Pagination } from "@/components/ui";
import { useMyBookmarks } from "@/hooks";
import { ROUTES } from "@/constants";

const BASE_PATH = ROUTES.PATIENT.FORUM;

export default function MyBookmarksPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useMyBookmarks({ page, pageSize: 12 });

  const posts = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 1;

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader title="My bookmarks" />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : posts.length === 0 ? (
        <Card className="p-10 text-center text-sm text-muted-foreground">You haven&apos;t bookmarked any posts yet.</Card>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {posts.map((post) => (
              <PostCard key={post.id} post={post} basePath={BASE_PATH} />
            ))}
          </div>
          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
    </div>
  );
}
