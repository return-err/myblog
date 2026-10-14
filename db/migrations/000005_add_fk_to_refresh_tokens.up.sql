ALTER TABLE refresh_tokens
ADD CONSTRAINT fk_user_refresh_tokens
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE CASCADE;