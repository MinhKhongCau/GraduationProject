"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus } from "lucide-react";
import { PostCard, PostFilters } from "@/components/forum";
import { Button, Spinner, Pagination } from "@/components/ui";
import { useForumPosts, useForumCategories, useForumTags, useDebounce } from "@/hooks";
import { ROUTES } from "@/constants";

const BASE_PATH = ROUTES.EXPERT.FORUM;

export default function ExpertForumPage() {
  const [search, setSearch] = useState("");
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [tag, setTag] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounce(search);

  const { data: categories = [] } = useForumCategories();
  const { data: tags = [] } = useForumTags();
  const { data, isLoading } = useForumPosts({
    search: debouncedSearch || undefined,
    categoryId: categoryId ?? undefined,
    tag: tag ?? undefined,
    page,
    pageSize: 12,
  });

  const posts = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 1;

  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Community</h1>
        <div className="flex gap-2">
          <Link href={ROUTES.EXPERT.FORUM_MY_BOOKMARKS}>
            <Button variant="outline" size="sm">
              My bookmarks
            </Button>
          </Link>
          <Link href={ROUTES.EXPERT.FORUM_NEW_POST}>
            <Button size="sm" className="flex items-center gap-1">
              <Plus className="h-4 w-4" /> New post
            </Button>
          </Link>
        </div>
      </div>

      <PostFilters
        search={search}
        onSearchChange={(v) => {
          setSearch(v);
          setPage(1);
        }}
        categoryId={categoryId}
        onCategoryChange={(v) => {
          setCategoryId(v);
          setPage(1);
        }}
        tag={tag}
        onTagChange={(v) => {
          setTag(v);
          setPage(1);
        }}
        categories={categories}
        tags={tags}
      />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : posts.length === 0 ? (
        <p className="text-sm text-muted-foreground">No posts found.</p>
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
