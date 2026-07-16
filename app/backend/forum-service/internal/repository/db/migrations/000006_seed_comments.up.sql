-- Seeds one root comment and one reply on a seeded post. Ids are assigned
-- explicitly (rather than left to the id BIGSERIAL default) so the ltree
-- `path` values can be written directly, producing the same shape the
-- runtime two-step insert-then-update algorithm (SPEC.md §3.1) would: a
-- root comment's path is its own id, a reply's path is its parent's path
-- plus its own id. This is safe here specifically because the table is
-- freshly created and empty at this point in the migration sequence.
-- (An earlier version of this migration tried to compute paths via chained
-- data-modifying CTEs — INSERT the row, then UPDATE it by its returned id
-- in a sibling CTE — but Postgres CTEs in one WITH clause share a single
-- snapshot, so an UPDATE re-scanning the target table can never see a row
-- a sibling CTE just inserted. The UPDATE ends up silently matching zero
-- rows.)
INSERT INTO comments (id, post_id, user_id, parent_id, path, content)
SELECT 1, id, 1004, NULL, '1'::ltree, 'Great write-up, this helped me set up my first container!'
FROM posts WHERE slug = 'getting-started-with-docker';

INSERT INTO comments (id, post_id, user_id, parent_id, path, content)
SELECT 2, id, 1005, 1, '1.2'::ltree, 'Agreed! The volume mounting section was especially clear.'
FROM posts WHERE slug = 'getting-started-with-docker';

SELECT setval(pg_get_serial_sequence('comments', 'id'), (SELECT MAX(id) FROM comments));

UPDATE posts SET comment_count = comment_count + 2
WHERE slug = 'getting-started-with-docker';
