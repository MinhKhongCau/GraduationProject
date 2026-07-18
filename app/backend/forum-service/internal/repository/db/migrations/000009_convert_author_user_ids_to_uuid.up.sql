-- forum-service originally stored author_id/user_id as BIGINT, disconnected
-- from the rest of MindCare where every account id (accountId/expertId/
-- patientId) is a UUID. The gateway forwards the JWT's accountId (a UUID)
-- as X-User-Id, so every authenticated write was rejected with 401 until
-- this migration (see internal/api/http/middleware/auth.go).
--
-- Seed rows (migrations 000004/000006/000008) used small integers
-- (1001-1005) as placeholder author/user ids, and there is no real account
-- to map them to, so they are deterministically encoded into the low bits of
-- a placeholder UUID (00000000-0000-0000-0000-xxxxxxxxxxxx) rather than
-- dropped, keeping the seed data structurally valid. Written as an inline
-- USING expression (not a CREATE FUNCTION) because the migration runner's
-- naive semicolon-based statement splitter (x-multi-statement) doesn't
-- understand dollar-quoted function bodies.
ALTER TABLE posts ALTER COLUMN author_id TYPE UUID USING (
    substr(lpad(to_hex(author_id), 32, '0'), 1, 8) || '-' ||
    substr(lpad(to_hex(author_id), 32, '0'), 9, 4) || '-' ||
    substr(lpad(to_hex(author_id), 32, '0'), 13, 4) || '-' ||
    substr(lpad(to_hex(author_id), 32, '0'), 17, 4) || '-' ||
    substr(lpad(to_hex(author_id), 32, '0'), 21, 12)
)::uuid;

ALTER TABLE comments ALTER COLUMN user_id TYPE UUID USING (
    substr(lpad(to_hex(user_id), 32, '0'), 1, 8) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 9, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 13, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 17, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 21, 12)
)::uuid;

ALTER TABLE post_likes ALTER COLUMN user_id TYPE UUID USING (
    substr(lpad(to_hex(user_id), 32, '0'), 1, 8) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 9, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 13, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 17, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 21, 12)
)::uuid;

ALTER TABLE post_bookmarks ALTER COLUMN user_id TYPE UUID USING (
    substr(lpad(to_hex(user_id), 32, '0'), 1, 8) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 9, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 13, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 17, 4) || '-' ||
    substr(lpad(to_hex(user_id), 32, '0'), 21, 12)
)::uuid;
