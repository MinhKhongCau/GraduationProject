INSERT INTO categories (name, slug, description) VALUES
    ('Programming', 'programming', 'Discussions about software development'),
    ('General Discussion', 'general-discussion', 'Anything and everything community-related'),
    ('Mental Health', 'mental-health', 'Support and discussion around mental wellbeing'),
    ('Wellness Tips', 'wellness-tips', 'Practical tips for a healthier daily routine');

INSERT INTO tags (name, slug) VALUES
    ('Docker', 'docker'),
    ('Go', 'golang'),
    ('Mindfulness', 'mindfulness'),
    ('Self-Care', 'self-care'),
    ('Career', 'career');

INSERT INTO posts (category_id, author_id, title, slug, summary, content) VALUES
    ((SELECT id FROM categories WHERE slug = 'programming'), 1001, 'Getting Started with Docker', 'getting-started-with-docker',
     'A quick intro to containers.', E'## Docker basics\n\nDocker packages an application and its dependencies into a portable container image.'),
    ((SELECT id FROM categories WHERE slug = 'programming'), 1002, 'Golang Concurrency Patterns', 'golang-concurrency-patterns',
     'Goroutines, channels, and common pitfalls.', E'## Concurrency in Go\n\nGoroutines are cheap, but misusing channels can still cause deadlocks.'),
    ((SELECT id FROM categories WHERE slug = 'general-discussion'), 1001, 'Welcome to the Community', 'welcome-to-the-community',
     'Say hello and introduce yourself!', E'Welcome! Feel free to introduce yourself below and tell us what brought you here.'),
    ((SELECT id FROM categories WHERE slug = 'mental-health'), 1003, 'Managing Exam Stress', 'managing-exam-stress',
     'A few techniques that helped me stay calm.', E'Exam season is stressful for everyone. Here are a few techniques that helped me stay grounded.'),
    ((SELECT id FROM categories WHERE slug = 'wellness-tips'), 1002, 'Five-Minute Mindfulness Practices', 'five-minute-mindfulness-practices',
     'Short practices you can fit into a busy day.', E'Even five minutes of mindfulness between tasks can reset your focus for the rest of the day.'),
    ((SELECT id FROM categories WHERE slug = 'wellness-tips'), 1003, 'Work-Life Balance Tips for Developers', 'work-life-balance-tips-for-developers',
     'Practical advice for avoiding burnout.', E'Burnout creeps up slowly. Setting clear boundaries around work hours is one of the best defenses.');

INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM (VALUES
    ('getting-started-with-docker', 'docker'),
    ('getting-started-with-docker', 'career'),
    ('golang-concurrency-patterns', 'golang'),
    ('golang-concurrency-patterns', 'career'),
    ('managing-exam-stress', 'mindfulness'),
    ('managing-exam-stress', 'self-care'),
    ('five-minute-mindfulness-practices', 'mindfulness'),
    ('five-minute-mindfulness-practices', 'self-care'),
    ('work-life-balance-tips-for-developers', 'career'),
    ('work-life-balance-tips-for-developers', 'self-care')
) AS link(post_slug, tag_slug)
JOIN posts p ON p.slug = link.post_slug
JOIN tags t ON t.slug = link.tag_slug;
