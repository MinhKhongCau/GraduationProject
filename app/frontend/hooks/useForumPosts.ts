"use client";

import { useApiQuery } from "./useApiQuery";
import { forumApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ListPostsParams } from "@/types";

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
