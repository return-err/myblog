CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_posts_title_trgm
ON posts
USING GIN (title gin_trgm_ops);