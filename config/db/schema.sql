CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tbl_users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    last_login_at TIMESTAMPTZ,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

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

CREATE TABLE IF NOT EXISTS tbl_analyses (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    sentence_id BIGINT NOT NULL REFERENCES tbl_sentences(id) ON DELETE CASCADE,
    explanation TEXT NOT NULL,
    vocabulary JSONB NOT NULL DEFAULT '[]',
    grammar_focus TEXT NOT NULL DEFAULT '',
    example TEXT NOT NULL DEFAULT '',
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tbl_sentences_created_at ON tbl_sentences(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tbl_users_email ON tbl_users(LOWER(email));
CREATE INDEX IF NOT EXISTS idx_tbl_users_deleted_at ON tbl_users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_users_uuid ON tbl_users(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_users_status_id ON tbl_users(status_id);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_deleted_at ON tbl_sentences(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_uuid ON tbl_sentences(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_status_id ON tbl_sentences(status_id);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_next_review_at ON tbl_sentences(next_review_at);
CREATE INDEX IF NOT EXISTS idx_tbl_sentences_category ON tbl_sentences(category);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_sentence_id ON tbl_analyses(sentence_id);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_deleted_at ON tbl_analyses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_uuid ON tbl_analyses(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_status_id ON tbl_analyses(status_id);
