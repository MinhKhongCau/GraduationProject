"use client";

import { useState } from "react";
import { PostCard } from "@/components/forum";
import { Spinner, Pagination } from "@/components/ui";
import { useMyBookmarks } from "@/hooks";
import { ROUTES } from "@/constants";

const BASE_PATH = ROUTES.EXPERT.FORUM;

export default function ExpertMyBookmarksPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useMyBookmarks({ page, pageSize: 12 });

  const posts = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 1;

  return (
    <div className="mx-auto max-w-6xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">My bookmarks</h1>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : posts.length === 0 ? (
        <p className="text-sm text-muted-foreground">You haven&apos;t bookmarked any posts yet.</p>
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
