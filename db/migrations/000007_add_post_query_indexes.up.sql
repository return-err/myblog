CREATE INDEX IF NOT EXISTS idx_posts_created_id
ON posts (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_posts_user_created_id
ON posts (user_id, created_at DESC, id DESC);