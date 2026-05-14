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

CREATE INDEX IF NOT EXISTS idx_tbl_sentences_created_at ON tbl_sentences(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_deleted_at ON tbl_sentences(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_uuid ON tbl_sentences(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_status_id ON tbl_sentences(status_id);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_next_review_at ON tbl_sentences(next_review_at);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_category ON tbl_sentences(category);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_tbl_sentences_status_id;
DROP INDEX IF EXISTS idx_tbl_sentences_category;
DROP INDEX IF EXISTS idx_tbl_sentences_next_review_at;
DROP INDEX IF EXISTS idx_tbl_sentences_uuid;
DROP INDEX IF EXISTS idx_tbl_sentences_deleted_at;
DROP INDEX IF EXISTS idx_tbl_sentences_created_at;
DROP TABLE IF EXISTS tbl_sentences;
