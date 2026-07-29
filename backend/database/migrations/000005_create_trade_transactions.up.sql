CREATE TABLE trade_transactions (
    id BIGSERIAL PRIMARY KEY,
    ticker TEXT NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell')),
    lot NUMERIC NOT NULL CHECK (lot > 0),
    price NUMERIC NOT NULL CHECK (price > 0),
    transaction_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trade_transactions_ticker ON trade_transactions(ticker);
CREATE INDEX idx_trade_transactions_date ON trade_transactions(transaction_date);
