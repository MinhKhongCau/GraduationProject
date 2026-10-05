ALTER TABLE posts ALTER COLUMN author_id TYPE BIGINT USING ('x' || right(replace(author_id::text, '-', ''), 16))::bit(64)::bigint;
ALTER TABLE comments ALTER COLUMN user_id TYPE BIGINT USING ('x' || right(replace(user_id::text, '-', ''), 16))::bit(64)::bigint;
ALTER TABLE post_likes ALTER COLUMN user_id TYPE BIGINT USING ('x' || right(replace(user_id::text, '-', ''), 16))::bit(64)::bigint;
ALTER TABLE post_bookmarks ALTER COLUMN user_id TYPE BIGINT USING ('x' || right(replace(user_id::text, '-', ''), 16))::bit(64)::bigint;
