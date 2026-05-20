-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tbl_sentences (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    text TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'extension',
    category TEXT NOT NULL DEFAULT 'general',
    review_count INTEGER NOT NULL DEFAULT 0,
    review_interval INTEGER NOT NULL DEFAULT 1,
    ease_factor DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    last_rating TEXT,
    last_reviewed_at TIMESTAMPTZ,
    next_review_at TIMESTAMPTZ,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS tbl_sentences;
