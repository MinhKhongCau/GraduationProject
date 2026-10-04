"use client";

import dynamic from "next/dynamic";
import { Badge, Modal, Spinner } from "@/components/ui";
import { useForumPost } from "@/hooks";

const ContentViewer = dynamic(() => import("@/components/ui/MDXEditor"), { ssr: false });

export interface PostPreviewModalProps {
  slug: string | null;
  onOpenChange: (open: boolean) => void;
}

export function PostPreviewModal({ slug, onOpenChange }: PostPreviewModalProps) {
  const { data: post, isLoading } = useForumPost(slug ?? "");

  return (
    <Modal open={!!slug} onOpenChange={onOpenChange} title={post?.title ?? "Xem bài viết"} size="2xl">
      {isLoading || !post ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="space-y-4">
          {post.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {post.tags.map((tag) => (
                <Badge key={tag} tone="primary">
                  #{tag}
                </Badge>
              ))}
            </div>
          )}
          <ContentViewer value={post.content} readOnly />
        </div>
      )}
    </Modal>
  );
}
