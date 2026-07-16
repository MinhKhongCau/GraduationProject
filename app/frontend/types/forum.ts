// forum-service uses Postgres int64 identifiers for posts/comments/tags/
// categories, unlike every other MindCare service's UUID strings — keep
// those as numbers, only stringify when building URL path segments.
// authorId/userId are the exception: they're the account's UUID (matching
// every other service's accountId), not a forum-internal numeric id.

export type PostStatus = "DRAFT" | "PUBLISHED" | "ARCHIVED";

export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  createdAt: string;
}

export interface Tag {
  id: number;
  name: string;
  slug: string;
}

export interface Post {
  id: number;
  title: string;
  slug: string;
  summary: string;
  thumbnailUrl: string;
  categoryId: number;
  authorId: string;
  status: PostStatus;
  viewCount: number;
  likeCount: number;
  bookmarkCount: number;
  commentCount: number;
  tags: string[];
  createdAt: string;
}

export interface PostDetail extends Post {
  content: string;
  updatedAt: string;
}

/** Note the singular `total` field — not `totalItems`/`totalPages` like
 * types/common.ts's PaginatedResponse<T>, which forum-service does not use. */
export interface PostListResult {
  items: Post[];
  page: number;
  pageSize: number;
  total: number;
}

export interface Comment {
  id: number;
  postId: number;
  userId: string;
  parentId: number | null;
  path: string;
  content: string | null; // null when deleted
  deleted: boolean;
  createdAt: string;
  replies: Comment[];
}

export interface LikeResult {
  postId: number;
  liked: boolean;
  likeCount: number;
}

export interface BookmarkResult {
  postId: number;
  bookmarked: boolean;
  bookmarkCount: number;
}

export interface CreatePostRequest {
  categoryId: number;
  title: string;
  summary?: string;
  content: string;
  thumbnailUrl?: string;
  status?: PostStatus;
  tags?: string[];
}

export type UpdatePostRequest = Partial<Omit<CreatePostRequest, "status">>;

export interface CreateCommentRequest {
  content: string;
  parentId?: number | null;
}

export interface ListPostsParams {
  page?: number;
  pageSize?: number;
  categoryId?: number;
  tag?: string;
  authorId?: string;
  status?: PostStatus;
  search?: string;
}
