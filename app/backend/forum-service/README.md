# Forum Service

The **Forum Service** is a microservice responsible for the community forum domain of the MindCare platform: categories, posts, tags, nested comments, likes, and bookmarks. See `SPEC.md` in this directory for the database design, business rules, and event contracts published to the Notification Service.

## Conventions

* **Base path:** `/api/v1/forum`
* **Content-Type:** `application/json`
* **Authentication:** the service trusts identity headers forwarded by the API Gateway (Kong) after JWT verification — it does not validate JWTs itself.
  ```
  X-User-Id: <numeric user id>
  X-User-Role: <CLIENT | EXPERT | ADMIN>
  ```
  Endpoints marked **Headers: Required** return `401` if these are missing.
* **Permissions legend:** `Public` (no auth needed), `Authenticated` (any logged-in user), `Owner` (the post/comment's author, or `ADMIN`), `ADMIN`.

## Folder Structure

```
forum-service/
├── cmd/
│   └── api/
│       └── main.go         # Entry point: initializes config, DB, router, and starts the server
├── configs/
│   ├── config.go           # Loads the .env file into a struct (Viper or GoDotEnv)
│   └── .env                # Environment file with DB/port settings (not committed to Git)
├── internal/                # All core code lives here so it can't be imported outside this module
│   ├── api/                 # Delivery / Transport layer
│   │   ├── http/            # HTTP REST layer (Gin, Fiber, or Echo)
│   │   │   ├── handlers/    # Receives requests, calls services, returns responses
│   │   │   └── router.go    # Defines API routes
│   │   └── dto/             # Data Transfer Objects: structs validating API request/response payloads
│   ├── app/                 # Business logic layer
│   │   ├── entity/          # Structs representing the system's core domain objects (Post, Comment, Tag, ...)
│   │   └── service/         # Interfaces and implementations for business logic
│   └── repository/          # Data layer
│       ├── dao/             # Data Access Objects: structs mapped directly to database tables
│       └── db/              # Database connection (PostgreSQL via GORM/sqlx) and migration runner
├── go.mod
└── go.sum
```

## API Documentation

### Health Check

* **`GET /health`**
  * **Description:** Returns the health status of the forum service. Used by the gateway's active upstream healthcheck.

### Categories

* **`GET /api/v1/forum/categories`**
  * **Description:** Lists all categories.
  * **Permissions:** `Public`
  * **Response:** Array of `{ id, name, slug, description, createdAt }`.

* **`GET /api/v1/forum/categories/{slug}`**
  * **Description:** Retrieves a single category by slug.
  * **Permissions:** `Public`
  * **Response:** `{ id, name, slug, description, createdAt }`.

* **`POST /api/v1/forum/categories`**
  * **Description:** Creates a new category.
  * **Permissions:** `ADMIN`
  * **Headers:** `X-User-Id`, `X-User-Role` (Required)
  * **Request Body:**
    ```json
    { "name": "Programming", "slug": "programming", "description": "Discussions about software development" }
    ```
  * **Response:** `{ id, name, slug, description, createdAt }`.

* **`PUT /api/v1/forum/categories/{id}`**
  * **Description:** Updates an existing category's name/slug/description.
  * **Permissions:** `ADMIN`
  * **Headers:** `X-User-Id`, `X-User-Role` (Required)
  * **Request Body:**
    ```json
    { "name": "Programming & Dev", "slug": "programming", "description": "Updated description" }
    ```
  * **Response:** `{ id, name, slug, description, createdAt }`.

* **`DELETE /api/v1/forum/categories/{id}`**
  * **Description:** Deletes a category. Fails if the category still has posts (`categories.posts` is `ON DELETE RESTRICT`).
  * **Permissions:** `ADMIN`
  * **Headers:** `X-User-Id`, `X-User-Role` (Required)
  * **Response:** `409 Conflict` with `{ "error": "CATEGORY_HAS_POSTS" }` if posts still reference it; `204 No Content` on success.

### Posts

* **`GET /api/v1/forum/posts`**
  * **Description:** Lists posts. Only `PUBLISHED` posts are returned unless the caller is the author or `ADMIN` and explicitly filters by `status`.
  * **Permissions:** `Public`
  * **Query params:** `categoryId`, `tag` (slug), `authorId`, `status`, `search` (matches title/summary), `page`, `pageSize`.
  * **Response:**
    ```json
    {
      "items": [
        {
          "id": 123, "title": "Getting Started with Docker", "slug": "getting-started-with-docker",
          "summary": "A quick intro to containers.", "thumbnailUrl": "https://.../thumb.jpg",
          "categoryId": 2, "authorId": 45, "status": "PUBLISHED",
          "viewCount": 320, "likeCount": 12, "bookmarkCount": 4, "commentCount": 7,
          "tags": ["docker", "devops"], "createdAt": "2026-07-01T10:00:00Z"
        }
      ],
      "page": 1, "pageSize": 20, "total": 57
    }
    ```

* **`GET /api/v1/forum/posts/{slug}`**
  * **Description:** Retrieves full post detail by slug and increments `view_count`. Returns `404` for a non-`PUBLISHED` post unless the caller is the author or `ADMIN`.
  * **Permissions:** `Public` for `PUBLISHED` posts; `Owner` for `DRAFT`/`ARCHIVED`
  * **Response:** Same shape as a list item, plus `"content": "<full body>"` and `"updatedAt"`.

* **`POST /api/v1/forum/posts`**
  * **Description:** Creates a new post. Tags supplied by name that don't yet exist are auto-created and attached; existing tags (matched by slug) are reused.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Request Body:**
    ```json
    {
      "categoryId": 2, "title": "Getting Started with Docker", "summary": "A quick intro to containers.",
      "content": "## Docker basics\n...", "thumbnailUrl": "https://.../thumb.jpg",
      "status": "PUBLISHED", "tags": ["docker", "devops"]
    }
    ```
  * **Response:** The created post (see detail shape above). Publishes `forum.post.created` if `status = PUBLISHED` (see `SPEC.md` §5).

* **`PUT /api/v1/forum/posts/{id}`**
  * **Description:** Updates a post's title/summary/content/thumbnail/category/tags.
  * **Permissions:** `Owner`
  * **Headers:** `X-User-Id` (Required)
  * **Request Body:** Same shape as create (all fields optional/partial).
  * **Response:** The updated post.

* **`DELETE /api/v1/forum/posts/{id}`**
  * **Description:** Soft-deletes a post by setting `deleted_at` (see `SPEC.md` §3.6). The post and its comments/likes/bookmarks/tag associations are left in the database but excluded from all read endpoints.
  * **Permissions:** `Owner`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `204 No Content`.

* **`PATCH /api/v1/forum/posts/{id}/status`**
  * **Description:** Transitions a post's status (`DRAFT → PUBLISHED`, or `PUBLISHED/DRAFT → ARCHIVED`). Publishes `forum.post.created` on the first transition into `PUBLISHED`.
  * **Permissions:** `Owner`
  * **Headers:** `X-User-Id` (Required)
  * **Request Body:**
    ```json
    { "status": "PUBLISHED" }
    ```
  * **Response:** The updated post.

### Tags

* **`GET /api/v1/forum/tags`**
  * **Description:** Lists all tags.
  * **Permissions:** `Public`
  * **Response:** Array of `{ id, name, slug }`.

* **`GET /api/v1/forum/tags/{slug}/posts`**
  * **Description:** Lists `PUBLISHED` posts associated with a tag. Same pagination/response shape as `GET /posts`.
  * **Permissions:** `Public`

### Comments

* **`GET /api/v1/forum/posts/{postId}/comments`**
  * **Description:** Retrieves the full comment tree for a post, ordered by `path` and assembled into nested `replies` arrays.
  * **Permissions:** `Public`
  * **Response:**
    ```json
    [
      {
        "id": 1, "userId": 45, "content": "Great post!", "path": "1", "createdAt": "2026-07-01T11:00:00Z",
        "replies": [
          { "id": 5, "userId": 9, "content": "Agreed!", "path": "1.5", "createdAt": "2026-07-01T11:05:00Z", "replies": [] }
        ]
      }
    ]
    ```

* **`POST /api/v1/forum/posts/{postId}/comments`**
  * **Description:** Creates a root comment (no `parentId`) or a reply (`parentId` set to an existing comment on the same post). Computes `path` from the new row's id per `SPEC.md` §3.1, and increments `posts.comment_count`. Publishes `forum.comment.created`.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Request Body:**
    ```json
    { "content": "Agreed!", "parentId": 1 }
    ```
  * **Response:** `{ id, postId, userId, parentId, path, content, createdAt }`.

* **`PUT /api/v1/forum/comments/{id}`**
  * **Description:** Edits a comment's content.
  * **Permissions:** `Owner`
  * **Headers:** `X-User-Id` (Required)
  * **Request Body:**
    ```json
    { "content": "Agreed, and here's why..." }
    ```
  * **Response:** The updated comment.

* **`DELETE /api/v1/forum/comments/{id}`**
  * **Description:** Soft-deletes a comment by setting `deleted_at` (see `SPEC.md` §3.6). The row and its `path`/`parent_id` are left intact so replies stay correctly linked; read endpoints render it as a placeholder (`content: null`, `"deleted": true`) instead of the original text. `posts.comment_count` is decremented by 1.
  * **Permissions:** `Owner`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `204 No Content`.

### Likes

* **`POST /api/v1/forum/posts/{postId}/like`**
  * **Description:** Likes a post. No-op (`200`, no duplicate row) if already liked, since `(post_id, user_id)` is a composite primary key. Increments `posts.like_count` and publishes `forum.post.liked` only on the first like.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `{ "postId": 123, "liked": true, "likeCount": 13 }`.

* **`DELETE /api/v1/forum/posts/{postId}/like`**
  * **Description:** Removes the caller's like. No-op if not currently liked. Decrements `posts.like_count`.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `{ "postId": 123, "liked": false, "likeCount": 12 }`.

### Bookmarks

* **`POST /api/v1/forum/posts/{postId}/bookmark`**
  * **Description:** Bookmarks a post for the caller. No-op if already bookmarked. Increments `posts.bookmark_count`.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `{ "postId": 123, "bookmarked": true, "bookmarkCount": 5 }`.

* **`DELETE /api/v1/forum/posts/{postId}/bookmark`**
  * **Description:** Removes the caller's bookmark. No-op if not currently bookmarked. Decrements `posts.bookmark_count`.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** `{ "postId": 123, "bookmarked": false, "bookmarkCount": 4 }`.

* **`GET /api/v1/forum/users/me/bookmarks`**
  * **Description:** Lists the caller's bookmarked posts, most recently bookmarked first.
  * **Permissions:** `Authenticated`
  * **Headers:** `X-User-Id` (Required)
  * **Response:** Same paginated shape as `GET /posts`.

## API Summary Table

| Resource | Endpoint | Method | Permission | Description |
|---|---|---|---|---|
| Health | `/health` | GET | Public | Service health check |
| Categories | `/api/v1/forum/categories` | GET | Public | List categories |
| Categories | `/api/v1/forum/categories/{slug}` | GET | Public | Get category by slug |
| Categories | `/api/v1/forum/categories` | POST | ADMIN | Create category |
| Categories | `/api/v1/forum/categories/{id}` | PUT | ADMIN | Update category |
| Categories | `/api/v1/forum/categories/{id}` | DELETE | ADMIN | Delete category (fails if it has posts) |
| Posts | `/api/v1/forum/posts` | GET | Public | List/search/filter posts |
| Posts | `/api/v1/forum/posts/{slug}` | GET | Public / Owner | Get post detail, increments view count |
| Posts | `/api/v1/forum/posts` | POST | Authenticated | Create post |
| Posts | `/api/v1/forum/posts/{id}` | PUT | Owner | Update post |
| Posts | `/api/v1/forum/posts/{id}` | DELETE | Owner | Soft-delete post (sets `deleted_at`) |
| Posts | `/api/v1/forum/posts/{id}/status` | PATCH | Owner | Change post status (DRAFT/PUBLISHED/ARCHIVED) |
| Tags | `/api/v1/forum/tags` | GET | Public | List tags |
| Tags | `/api/v1/forum/tags/{slug}/posts` | GET | Public | List posts by tag |
| Comments | `/api/v1/forum/posts/{postId}/comments` | GET | Public | Get nested comment tree |
| Comments | `/api/v1/forum/posts/{postId}/comments` | POST | Authenticated | Create root comment or reply |
| Comments | `/api/v1/forum/comments/{id}` | PUT | Owner | Edit comment |
| Comments | `/api/v1/forum/comments/{id}` | DELETE | Owner | Soft-delete comment (sets `deleted_at`, replies stay linked) |
| Likes | `/api/v1/forum/posts/{postId}/like` | POST | Authenticated | Like a post |
| Likes | `/api/v1/forum/posts/{postId}/like` | DELETE | Authenticated | Unlike a post |
| Bookmarks | `/api/v1/forum/posts/{postId}/bookmark` | POST | Authenticated | Bookmark a post |
| Bookmarks | `/api/v1/forum/posts/{postId}/bookmark` | DELETE | Authenticated | Remove bookmark |
| Bookmarks | `/api/v1/forum/users/me/bookmarks` | GET | Authenticated | List my bookmarked posts |
