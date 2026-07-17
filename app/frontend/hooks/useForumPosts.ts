"use client";

import { useApiQuery } from "./useApiQuery";
import { forumApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ListPostsParams, Post, PostStatus } from "@/types";

export function useForumPosts(params: ListPostsParams = {}) {
  return useApiQuery({
    queryKey: QUERY_KEYS.forumPosts(params as Record<string, unknown>),
    queryFn: () => forumApi.listPosts(params),
  });
}

export function useForumPost(slug: string) {
  return useApiQuery({
    queryKey: QUERY_KEYS.forumPost(slug),
    queryFn: () => forumApi.getPostBySlug(slug),
    enabled: !!slug,
  });
}

export function useForumCategories() {
  return useApiQuery({
    queryKey: QUERY_KEYS.forumCategories(),
    queryFn: forumApi.listCategories,
  });
}

export function useForumTags() {
  return useApiQuery({
    queryKey: QUERY_KEYS.forumTags(),
    queryFn: forumApi.listTags,
  });
}

export function useForumComments(postId: number) {
  return useApiQuery({
    queryKey: QUERY_KEYS.forumComments(postId),
    queryFn: () => forumApi.getCommentTree(postId),
    enabled: !!postId,
  });
}

export function useMyBookmarks(params: { page?: number; pageSize?: number } = {}) {
  return useApiQuery({
    queryKey: QUERY_KEYS.myBookmarks(),
    queryFn: () => forumApi.listMyBookmarks(params),
  });
}

const MY_POSTS_STATUSES: PostStatus[] = ["PUBLISHED", "DRAFT", "ARCHIVED"];

/**
 * forum-service's List endpoint always defaults to status=PUBLISHED when no
 * status is given, even for the post's own author — so seeing every one of
 * "my" posts (drafts + blocked/archived included) requires one request per
 * status, merged client-side.
 */
export function useMyPosts(userId: string | undefined) {
  return useApiQuery({
    queryKey: QUERY_KEYS.myPosts(userId),
    queryFn: async (): Promise<Post[]> => {
      const results = await Promise.all(
        MY_POSTS_STATUSES.map((status) => forumApi.listPosts({ authorId: userId, status, pageSize: 100 }))
      );
      return results
        .flatMap((result) => result.items)
        .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
    },
    enabled: !!userId,
  });
}
