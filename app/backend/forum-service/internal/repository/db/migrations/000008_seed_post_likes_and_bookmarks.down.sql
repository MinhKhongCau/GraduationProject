DELETE FROM post_likes
WHERE post_id IN (SELECT id FROM posts WHERE slug IN (
    'getting-started-with-docker', 'golang-concurrency-patterns', 'five-minute-mindfulness-practices'
))
AND user_id IN (1001, 1002, 1003, 1004, 1005);

DELETE FROM post_bookmarks
WHERE post_id IN (SELECT id FROM posts WHERE slug IN (
    'getting-started-with-docker', 'managing-exam-stress'
))
AND user_id IN (1001, 1002, 1005);

UPDATE posts SET like_count = 0 WHERE slug IN (
    'getting-started-with-docker', 'golang-concurrency-patterns', 'five-minute-mindfulness-practices'
);
UPDATE posts SET bookmark_count = 0 WHERE slug IN (
    'getting-started-with-docker', 'managing-exam-stress'
);
