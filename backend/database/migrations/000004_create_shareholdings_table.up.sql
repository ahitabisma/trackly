CREATE TABLE shareholdings (
    id BIGSERIAL PRIMARY KEY,

    company_id BIGINT NOT NULL,
    investor_id BIGINT NOT NULL,

    date DATE NOT NULL,

    holdings_scripless BIGINT DEFAULT 0,
    holdings_scrip BIGINT DEFAULT 0,
    total_holding_shares BIGINT NOT NULL,

    percentage NUMERIC(6,2) NOT NULL,

    source VARCHAR(50) NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_company
        FOREIGN KEY (company_id)
        REFERENCES companies(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_investor
        FOREIGN KEY (investor_id)
        REFERENCES investors(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_snapshot
        UNIQUE (company_id, investor_id, date)
);

-- query per company + tanggal
CREATE INDEX idx_shareholdings_company_date 
ON shareholdings(company_id, date);

-- query top holder
CREATE INDEX idx_shareholdings_company_percentage 
ON shareholdings(company_id, percentage DESC);

-- query investor history
CREATE INDEX idx_shareholdings_investor 
ON shareholdings(investor_id);