-- Drop post_media junction table first (due to foreign key constraints)
DROP TABLE IF EXISTS post_media;

-- Drop posts table after post_media is removed
DROP TABLE IF EXISTS posts;