CREATE TABLE
    api_tokens (
        id BIGSERIAL PRIMARY KEY,
        token_hash TEXT NOT NULL UNIQUE,
        role TEXT NOT NULL,
        label TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        revoked_at TIMESTAMPTZ
    );

CREATE INDEX idx_api_tokens_hash ON api_tokens (token_hash)
WHERE
    revoked_at IS NULL;