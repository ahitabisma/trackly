CREATE TABLE investors (
    id BIGSERIAL PRIMARY KEY,

    canonical_name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,

    investor_type VARCHAR(20) NULL, -- ID / CP / etc
    local_foreign CHAR(1) NULL, -- L / F

    nationality VARCHAR(100) NULL,
    domicile VARCHAR(100) NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_investors_normalized ON investors(normalized_name);

-- Tabel alias untuk semua variasi nama yang ketemu di raw data
CREATE TABLE investor_aliases (
    id BIGSERIAL PRIMARY KEY,
    investor_id BIGINT NOT NULL REFERENCES investors(id) ON DELETE CASCADE,
    alias_name VARCHAR(255) NOT NULL,       -- nama asli dari sumber data
    normalized_alias VARCHAR(255) NOT NULL, -- versi normalized-nya
    source VARCHAR(50) NULL,               -- dari mana alias ini berasal
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(normalized_alias)
);

CREATE INDEX idx_aliases_investor ON investor_aliases(investor_id);
CREATE INDEX idx_aliases_normalized ON investor_aliases(normalized_alias);