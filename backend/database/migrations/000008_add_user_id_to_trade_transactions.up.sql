ALTER TABLE trade_transactions ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_trade_transactions_user_id ON trade_transactions(user_id);
