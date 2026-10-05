"use client";

import { useState } from "react";
import { Eye, Ban, Trash2 } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Card, Button, Spinner, Pagination, PageHeader, Select, Badge, type BadgeTone } from "@/components/ui";
import { useApiQuery, useApiMutation, useForumAuthorName } from "@/hooks";
import { forumApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { PostPreviewModal } from "./component/PostPreviewModal";
import type { Post, PostStatus } from "@/types";

const PAGE_SIZE = 10;

const STATUS_OPTIONS: { value: PostStatus; label: string }[] = [
  { value: "PUBLISHED", label: "Đã đăng" },
  { value: "DRAFT", label: "Bản nháp" },
  { value: "ARCHIVED", label: "Đã chặn" },
];

const STATUS_BADGE: Record<PostStatus, BadgeTone> = {
  PUBLISHED: "success",
  DRAFT: "neutral",
  ARCHIVED: "danger",
};

function AuthorCell({ authorId }: { authorId: string }) {
  return <>{useForumAuthorName(authorId)}</>;
}

export default function AdminForumPostsPage() {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [status, setStatus] = useState<PostStatus>("PUBLISHED");
  const [page, setPage] = useState(1);
  const [previewSlug, setPreviewSlug] = useState<string | null>(null);

  const params = { status, page, pageSize: PAGE_SIZE };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.forumPosts(params),
    queryFn: () => forumApi.listPosts(params),
  });

  const changeStatusMutation = useApiMutation({
    mutationFn: ({ postId, status }: { postId: number; status: PostStatus }) =>
      forumApi.changePostStatus(postId, status),
    onSuccess: () => {
      showSuccess("Cập nhật trạng thái bài viết thành công!");
      queryClient.invalidateQueries({ queryKey: ["forum", "posts"] });
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (postId: number) => forumApi.deletePost(postId),
    onSuccess: () => {
      showSuccess("Xóa bài viết thành công!");
      queryClient.invalidateQueries({ queryKey: ["forum", "posts"] });
    },
  });

  const posts = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 1;

  function handleBlock(post: Post) {
    if (confirm(`Chặn bài viết "${post.title}"? Hành động này không thể hoàn tác.`)) {
      changeStatusMutation.mutate({ postId: post.id, status: "ARCHIVED" });
    }
  }

  function handleDelete(post: Post) {
    if (confirm(`Xóa bài viết "${post.title}"? Hành động này không thể hoàn tác.`)) {
      deleteMutation.mutate(post.id);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý bài viết cộng đồng"
        description="Kiểm duyệt, chặn hoặc xóa bài viết trên diễn đàn."
        actions={
          <Select
            aria-label="Lọc theo trạng thái"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value as PostStatus);
              setPage(1);
            }}
            className="w-full sm:w-48"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </Select>
        }
      />

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="border-b border-border bg-surface text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="px-4 py-3">Tiêu đề</th>
                  <th className="px-4 py-3">Tác giả</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3">Lượt thích</th>
                  <th className="px-4 py-3">Bình luận</th>
                  <th className="px-4 py-3 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {posts.map((post) => (
                  <tr key={post.id} className="transition-colors hover:bg-surface/60">
                    <td className="max-w-xs truncate px-4 py-3 font-medium text-foreground">{post.title}</td>
                    <td className="px-4 py-3 text-muted-foreground">
                      <AuthorCell authorId={post.authorId} />
                    </td>
                    <td className="px-4 py-3">
                      <Badge tone={STATUS_BADGE[post.status]}>
                        {STATUS_OPTIONS.find((o) => o.value === post.status)?.label ?? post.status}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{post.likeCount}</td>
                    <td className="px-4 py-3 text-muted-foreground">{post.commentCount}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setPreviewSlug(post.slug)}
                          className="text-primary hover:bg-primary-soft"
                        >
                          <Eye className="h-3.5 w-3.5" /> Xem
                        </Button>
                        {post.status !== "ARCHIVED" && (
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={changeStatusMutation.isPending}
                            onClick={() => handleBlock(post)}
                            className="text-danger hover:bg-danger-soft"
                          >
                            <Ban className="h-3.5 w-3.5" /> Chặn
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={deleteMutation.isPending}
                          onClick={() => handleDelete(post)}
                          className="text-danger hover:bg-danger-soft"
                        >
                          <Trash2 className="h-3.5 w-3.5" /> Xóa
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {posts.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không tìm thấy bài viết nào.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        </Card>
      )}

      <PostPreviewModal slug={previewSlug} onOpenChange={(open) => !open && setPreviewSlug(null)} />
    </div>
  );
}
