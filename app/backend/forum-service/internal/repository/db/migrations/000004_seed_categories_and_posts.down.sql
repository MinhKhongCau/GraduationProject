DELETE FROM post_tags WHERE post_id IN (
    SELECT id FROM posts WHERE slug IN (
        'getting-started-with-docker',
        'golang-concurrency-patterns',
        'welcome-to-the-community',
        'managing-exam-stress',
        'five-minute-mindfulness-practices',
        'work-life-balance-tips-for-developers'
    )
);

DELETE FROM posts WHERE slug IN (
    'getting-started-with-docker',
    'golang-concurrency-patterns',
    'welcome-to-the-community',
    'managing-exam-stress',
    'five-minute-mindfulness-practices',
    'work-life-balance-tips-for-developers'
);

DELETE FROM tags WHERE slug IN ('docker', 'golang', 'mindfulness', 'self-care', 'career');

DELETE FROM categories WHERE slug IN ('programming', 'general-discussion', 'mental-health', 'wellness-tips');
