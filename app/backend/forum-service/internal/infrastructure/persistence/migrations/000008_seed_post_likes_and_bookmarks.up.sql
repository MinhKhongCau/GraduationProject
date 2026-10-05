INSERT INTO post_likes (post_id, user_id)
SELECT p.id, u.user_id
FROM (VALUES
    ('getting-started-with-docker', 1001),
    ('getting-started-with-docker', 1002),
    ('getting-started-with-docker', 1003),
    ('getting-started-with-docker', 1004),
    ('golang-concurrency-patterns', 1001),
    ('golang-concurrency-patterns', 1003),
    ('five-minute-mindfulness-practices', 1002),
    ('five-minute-mindfulness-practices', 1004),
    ('five-minute-mindfulness-practices', 1005)
) AS u(post_slug, user_id)
JOIN posts p ON p.slug = u.post_slug;

INSERT INTO post_bookmarks (post_id, user_id)
SELECT p.id, u.user_id
FROM (VALUES
    ('getting-started-with-docker', 1002),
    ('getting-started-with-docker', 1005),
    ('managing-exam-stress', 1001)
) AS u(post_slug, user_id)
JOIN posts p ON p.slug = u.post_slug;

UPDATE posts SET like_count = 4 WHERE slug = 'getting-started-with-docker';
UPDATE posts SET like_count = 2 WHERE slug = 'golang-concurrency-patterns';
UPDATE posts SET like_count = 3 WHERE slug = 'five-minute-mindfulness-practices';

UPDATE posts SET bookmark_count = 2 WHERE slug = 'getting-started-with-docker';
UPDATE posts SET bookmark_count = 1 WHERE slug = 'managing-exam-stress';
