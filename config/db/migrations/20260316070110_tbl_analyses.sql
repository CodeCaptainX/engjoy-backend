-- +goose Up
-- +goose StatementBegin
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

CREATE INDEX IF NOT EXISTS idx_tbl_analyses_sentence_id ON tbl_analyses(sentence_id);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_deleted_at ON tbl_analyses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_uuid ON tbl_analyses(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_analyses_status_id ON tbl_analyses(status_id);
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_tbl_analyses_status_id;
DROP INDEX IF EXISTS idx_tbl_analyses_uuid;
DROP INDEX IF EXISTS idx_tbl_analyses_deleted_at;
DROP INDEX IF EXISTS idx_tbl_analyses_sentence_id;
DROP TABLE IF EXISTS tbl_analyses;
