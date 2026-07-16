DELETE FROM comments WHERE id IN (1, 2);

UPDATE posts SET comment_count = comment_count - 2
WHERE slug = 'getting-started-with-docker';
