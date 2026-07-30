DROP INDEX IF EXISTS idx_trade_transactions_user_id;

ALTER TABLE trade_transactions DROP COLUMN IF EXISTS user_id;
