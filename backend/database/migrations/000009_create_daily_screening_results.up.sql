CREATE TABLE IF NOT EXISTS daily_screening_results (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_date   DATE NOT NULL,
    ticker      VARCHAR(10) NOT NULL,
    rank        SMALLINT NOT NULL CHECK (rank BETWEEN 1 AND 10),
    score       NUMERIC(5,2) NOT NULL,
    confidence  VARCHAR(10) NOT NULL DEFAULT 'low',
    overall     VARCHAR(10) NOT NULL DEFAULT 'neutral',
    avg_volume  NUMERIC(20,0),
    trading_plan   JSONB,
    ai_insight     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scan_date, ticker)
);

CREATE INDEX idx_screening_date ON daily_screening_results(scan_date);
CREATE INDEX idx_screening_rank ON daily_screening_results(scan_date, rank);
