ALTER TABLE users ADD COLUMN telegram_chat_id BIGINT;
CREATE INDEX idx_users_telegram_chat_id ON users(telegram_chat_id);
