# Forum Service — Technical Specification

## 1. Overview

The **Forum Service** owns all community-discussion data and business logic for the MindCare platform: categories, posts, tags, nested comments, likes, and bookmarks. It is one service in a larger event-driven microservices architecture:

* **Forum Service** — source of truth for `categories`, `posts`, `tags`, `post_tags`, `comments`, `post_likes`, `post_bookmarks`. Owns its own PostgreSQL database and exposes a REST API (see `README.md` in this directory).
* **Notification Service** — a separate service that consumes domain events published by the Forum Service (new post, new comment/reply, new like) from the shared message broker and delivers real-time notifications to users. The Forum Service never calls the Notification Service directly — it only publishes events (see [§5 Events Published](#5-events-published)).
* **User Service** — owns user identity. The Forum Service does not store user profile data; `author_id` / `user_id` columns are opaque foreign references resolved by the caller (frontend / BFF) or by other services, not joined inside this service.

## 2. Database Design (PostgreSQL)

### 2.1 Categories, Posts & Tags

```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL, -- e.g. "Programming", "General Discussion"
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

```sql
CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    author_id BIGINT NOT NULL, -- user id, managed by User Service
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    summary TEXT, -- short description shown in list views
    content TEXT NOT NULL, -- post body (raw Text/Markdown/HTML)
    thumbnail_url VARCHAR(512),
    status VARCHAR(50) DEFAULT 'PUBLISHED', -- DRAFT, PUBLISHED, ARCHIVED
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,       -- counter cache
    bookmark_count INT DEFAULT 0,   -- counter cache
    comment_count INT DEFAULT 0,    -- counter cache
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE -- soft delete marker; NULL = active, non-NULL = deleted
);

CREATE INDEX idx_posts_category ON posts(category_id);
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status_created ON posts(status, created_at DESC) WHERE deleted_at IS NULL;
```

```sql
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL, -- e.g. "Java", "Docker"
    slug VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE post_tags (
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_post_tags_tag ON post_tags(tag_id);
```

### 2.2 Nested Comments

Hierarchy is tracked via `parent_id` plus a dot-separated `path` for fast subtree queries.

```sql
CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    parent_id BIGINT REFERENCES comments(id) ON DELETE SET NULL,
    path ltree NOT NULL, -- hierarchical path, e.g. "1", "1.5", "1.5.12"
    content TEXT NOT NULL, -- supports UTF-8 special characters / emoji
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE -- soft delete marker; NULL = active, non-NULL = deleted
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_path ON comments (path varchar_pattern_ops);
```

### 2.3 Likes & Bookmarks

```sql
CREATE TABLE post_likes (
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id) -- prevents duplicate likes
);

CREATE INDEX idx_post_likes_post ON post_likes(post_id);
```

```sql
CREATE TABLE post_bookmarks (
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_post_bookmarks_user ON post_bookmarks(user_id);
```

### 2.4 Relationship Summary

| Relation | Cardinality | ON DELETE behavior |
|---|---|---|
| `categories` → `posts` | 1 – N | `RESTRICT` — a category with existing posts cannot be deleted |
| `posts` ↔ `tags` (via `post_tags`) | N – N | `CASCADE` on both sides — deleting a post or tag removes the join rows |
| `posts` → `comments` | 1 – N | `CASCADE` — deleting a post deletes its comments |
| `comments` → `comments` (self, `parent_id`) | 1 – N | `SET NULL` — deleting a parent comment orphans replies rather than cascading the delete (their `path` prefix still identifies the original thread) |
| `posts` ↔ `users` (via `post_likes`) | N – N | `CASCADE` from `posts`; no FK to a `users` table (User Service owns that) |
| `posts` ↔ `users` (via `post_bookmarks`) | N – N | `CASCADE` from `posts`; no FK to a `users` table |

> **Note:** as of §3.6, the everyday API `DELETE` on `posts`/`comments` performs a **soft delete** (sets `deleted_at`) rather than issuing a real `DELETE FROM`. The `ON DELETE CASCADE`/`RESTRICT`/`SET NULL` behavior above still governs genuine row deletion (e.g. a data-retention/cleanup job or hard category removal), just not the standard API delete flow.

## 3. Business Rules

### 3.1 Nested comment `path` computation

`path` cannot be computed before insert because it needs the new row's own generated `id`. The service must:

1. Insert the comment row with a placeholder `path` (e.g. `'0'` or the string form of `parent_id` alone) to obtain the generated `id`.
2. Compute the final `path`:
   * Root comment (`parent_id IS NULL`): `path = id::text`.
   * Reply (`parent_id = P`): `path = parent.path || '.' || id::text`.
3. `UPDATE comments SET path = <computed> WHERE id = <new id>` in the same transaction as step 1.

Fetching a full subtree below comment `1.5`:

```sql
SELECT * FROM comments WHERE post_id = ? AND path LIKE '1.5.%' ORDER BY path ASC;
```

Fetching a post's entire comment tree (for building a nested response) is the same query without a prefix filter, ordered by `path ASC`, then assembled client-side/server-side into a tree using `parent_id`.

### 3.2 Counter caches

`posts.like_count`, `posts.bookmark_count`, `posts.comment_count`, and `posts.view_count` are denormalized counters, not computed via `COUNT(*)` on read. Each mutating action increments/decrements the relevant counter on `posts` in the same transaction as the underlying insert/delete:

* Comment created → `comment_count += 1`. Comment soft-deleted → `comment_count -= 1` (only for the deleted row itself; replies are untouched, see §3.6).
* Like inserted → `like_count += 1`. Like removed → `like_count -= 1`.
* Bookmark inserted → `bookmark_count += 1`. Bookmark removed → `bookmark_count -= 1`.
* Post viewed (detail endpoint) → `view_count += 1`. This should not be blocked by the same request/response cycle if high write volume is a concern (e.g. can be a fire-and-forget increment), but is specified here as synchronous for correctness.

### 3.3 Like / bookmark idempotency

`post_likes` and `post_bookmarks` use `(post_id, user_id)` as a composite primary key specifically to reject duplicate like/bookmark spam at the database level. The API models this as two explicit actions — `POST .../like` (create, `409`/no-op if already liked) and `DELETE .../like` (remove, no-op if not liked) — rather than a single toggle endpoint, so clients always know the resulting state from the HTTP method alone.

### 3.4 Post status lifecycle

`status` moves `DRAFT → PUBLISHED → ARCHIVED`. Only `PUBLISHED` posts are returned by public list/detail endpoints; `DRAFT` and `ARCHIVED` posts are visible only to their author or an `ADMIN`. There is no defined transition back from `ARCHIVED` to `PUBLISHED` in this spec — archiving is a terminal, author-chosen state, but an archived post remains visible to its author/`ADMIN` and is distinct from a deleted one (see §3.6).

### 3.5 Category deletion

Because `posts.category_id` uses `ON DELETE RESTRICT`, a category with at least one post cannot be deleted; the API must surface this as a `409 Conflict`, not a raw database error.

### 3.6 Soft delete (posts & comments)

Both `posts` and `comments` carry a `deleted_at` column. The everyday API delete flow sets this timestamp instead of issuing a hard `DELETE FROM`:

* **Posts** — `DELETE /api/v1/forum/posts/{id}` sets `posts.deleted_at = now()`. All read endpoints (list, detail, tag listing, bookmarks) filter `WHERE deleted_at IS NULL`, so a soft-deleted post disappears even for its own author (no "trash" view is specified). Child rows (`comments`, `post_likes`, `post_bookmarks`, `post_tags`) are left in the database untouched — they simply become unreachable because the parent post no longer surfaces through the API.
* **Comments** — `DELETE /api/v1/forum/comments/{id}` sets `comments.deleted_at = now()`. Because the row itself is kept (not removed, and `parent_id`/`path` are untouched), replies underneath it remain correctly linked and are unaffected. Read endpoints render a soft-deleted comment as a placeholder (e.g. `content: null`, `"deleted": true`) instead of its original text, preserving thread structure — this supersedes the `ON DELETE SET NULL` orphaning described in §2.4, which only applies to a genuine hard delete, not this flow. `posts.comment_count` is still decremented when a comment is soft-deleted (§3.2).
* The `ON DELETE CASCADE`/`RESTRICT`/`SET NULL` foreign-key behavior defined in §2.1–§2.3 remains in the schema for actual hard deletes (e.g. GDPR erasure, category removal) — it does not fire during the normal soft-delete API flow above.

## 4. Identity & Authorization Model

The Forum Service does not validate JWTs itself. Consistent with the rest of this platform's services (`auth-service`, `payment-service`, `booking-service`, `profile-service`), it trusts identity headers injected by the Kong API Gateway after the gateway has already verified the end user's JWT via the `mindcare-auth` plugin:

| Header | Meaning | Used for |
|---|---|---|
| `X-User-Id` | Authenticated user's numeric id | Populates `author_id` / `user_id` on writes; used for "my bookmarks", "is this my post" ownership checks |
| `X-User-Role` | Authenticated user's role (e.g. `CLIENT`, `EXPERT`, `ADMIN`) | Gates `ADMIN`-only actions (category management) |

Endpoints that don't require identity (public reads) are reachable without these headers. Endpoints that do require them return `401 Unauthorized` if the headers are absent — the gateway is expected to have already rejected unauthenticated requests to protected routes, so a missing header at the service layer indicates a misconfigured route rather than a normal client error path.

## 5. Events Published

The Forum Service publishes domain events to the shared message broker (RabbitMQ, already provisioned in this platform's infrastructure though not yet consumed by any live service) so the Notification Service can deliver real-time notifications without the Forum Service knowing anything about delivery channels (push, email, in-app).

| Event | Trigger | Payload |
|---|---|---|
| `forum.post.created` | A post transitions to `PUBLISHED` (on create, or on a DRAFT→PUBLISHED status change) | `{ "postId": 123, "authorId": 45, "categoryId": 2, "title": "...", "publishedAt": "..." }` |
| `forum.comment.created` | A new comment or reply is created | `{ "commentId": 12, "postId": 123, "userId": 45, "postAuthorId": 7, "parentId": 5, "parentAuthorId": 9 }` — `parentId`/`parentAuthorId` are `null` for root comments; the Notification Service uses `parentAuthorId` to notify the specific comment being replied to, and `postAuthorId` to notify the post's author, without querying the Forum Service back |
| `forum.post.liked` | A like is created (not on unlike) | `{ "postId": 123, "userId": 45, "postAuthorId": 7 }` |

Each event carries enough denormalized context (`postAuthorId`, `parentAuthorId`) for the Notification Service to act without a callback into the Forum Service, keeping the two services decoupled.

## 6. Infrastructure Fit

* **Database**: own logical PostgreSQL database (env var `FORUM_DB_NAME`, following the `PROFILE_DB_NAME` / `PAYMENT_DB_NAME` / `BOOKING_DB_NAME` naming convention used by sibling services), added to the shared `postgres-db` container.
* **Health check**: exposes `GET /health` for Kong's active upstream healthcheck, matching every other service in this platform.
* **Gateway routing**: a public Kong route under `/api/v1/forum`, with the `mindcare-auth` plugin attached to the subset of routes that require `X-User-Id`/`X-User-Role` (writes, bookmarks) but not to public read routes (category/post/tag listing).
