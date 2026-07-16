"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useApiMutation } from "./useApiMutation";
import { forumApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { CreateCommentRequest, CreatePostRequest } from "@/types";

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useApiMutation({
    mutationFn: (payload: CreatePostRequest) => forumApi.createPost(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["forum", "posts"] }),
  });
}

export function useCreateComment(postId: number) {
  const queryClient = useQueryClient();
  return useApiMutation({
    mutationFn: (payload: CreateCommentRequest) => forumApi.createComment(postId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: QUERY_KEYS.forumComments(postId) }),
  });
}

export function useToggleLike(postId: number) {
  const queryClient = useQueryClient();
  return useApiMutation({
    mutationFn: (currentlyLiked: boolean) =>
      currentlyLiked ? forumApi.unlikePost(postId) : forumApi.likePost(postId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["forum", "post"] }),
  });
}

export function useToggleBookmark(postId: number) {
  const queryClient = useQueryClient();
  return useApiMutation({
    mutationFn: (currentlyBookmarked: boolean) =>
      currentlyBookmarked ? forumApi.removeBookmark(postId) : forumApi.bookmarkPost(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["forum", "post"] });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myBookmarks() });
    },
  });
}
