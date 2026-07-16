import { forumClient } from "./http/instances";
import { FORUM_ENDPOINTS } from "@/constants/api";
import type {
  Category,
  Tag,
  PostDetail,
  PostListResult,
  Comment,
  LikeResult,
  BookmarkResult,
  CreatePostRequest,
  UpdatePostRequest,
  CreateCommentRequest,
  ListPostsParams,
} from "@/types";

export async function listCategories(): Promise<Category[]> {
  const res = await forumClient.get<Category[]>(FORUM_ENDPOINTS.CATEGORIES);
  return res.data;
}

export async function listTags(): Promise<Tag[]> {
  const res = await forumClient.get<Tag[]>(FORUM_ENDPOINTS.TAGS);
  return res.data;
}

export async function listPosts(params: ListPostsParams = {}): Promise<PostListResult> {
  const res = await forumClient.get<PostListResult>(FORUM_ENDPOINTS.POSTS, { params });
  return res.data;
}

export async function listPostsByTag(slug: string, params: ListPostsParams = {}): Promise<PostListResult> {
  const res = await forumClient.get<PostListResult>(FORUM_ENDPOINTS.TAG_POSTS(slug), { params });
  return res.data;
}

/** GET /posts/{slug} — see FORUM_ENDPOINTS.POST's slug-vs-id note. */
export async function getPostBySlug(slug: string): Promise<PostDetail> {
  const res = await forumClient.get<PostDetail>(FORUM_ENDPOINTS.POST(slug));
  return res.data;
}

export async function createPost(payload: CreatePostRequest): Promise<PostDetail> {
  const res = await forumClient.post<PostDetail>(FORUM_ENDPOINTS.POSTS, payload);
  return res.data;
}

/** PUT /posts/{id} — numeric post id, not the slug. */
export async function updatePost(postId: number, payload: UpdatePostRequest): Promise<PostDetail> {
  const res = await forumClient.put<PostDetail>(FORUM_ENDPOINTS.POST(postId), payload);
  return res.data;
}

export async function getCommentTree(postId: number): Promise<Comment[]> {
  const res = await forumClient.get<Comment[]>(FORUM_ENDPOINTS.POST_COMMENTS(postId));
  return res.data;
}

export async function createComment(postId: number, payload: CreateCommentRequest): Promise<Comment> {
  const res = await forumClient.post<Comment>(FORUM_ENDPOINTS.POST_COMMENTS(postId), payload);
  return res.data;
}

export async function likePost(postId: number): Promise<LikeResult> {
  const res = await forumClient.post<LikeResult>(FORUM_ENDPOINTS.POST_LIKE(postId));
  return res.data;
}

export async function unlikePost(postId: number): Promise<LikeResult> {
  const res = await forumClient.delete<LikeResult>(FORUM_ENDPOINTS.POST_LIKE(postId));
  return res.data;
}

export async function bookmarkPost(postId: number): Promise<BookmarkResult> {
  const res = await forumClient.post<BookmarkResult>(FORUM_ENDPOINTS.POST_BOOKMARK(postId));
  return res.data;
}

export async function removeBookmark(postId: number): Promise<BookmarkResult> {
  const res = await forumClient.delete<BookmarkResult>(FORUM_ENDPOINTS.POST_BOOKMARK(postId));
  return res.data;
}

export async function listMyBookmarks(params: { page?: number; pageSize?: number } = {}): Promise<PostListResult> {
  const res = await forumClient.get<PostListResult>(FORUM_ENDPOINTS.MY_BOOKMARKS, { params });
  return res.data;
}

