-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_sentences_reactions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    "order" INTEGER NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL REFERENCES tbl_users(id),
    sentence_id BIGINT NOT NULL REFERENCES tbl_sentences(id),
    reaction_type VARCHAR(50) NOT NULL DEFAULT 'heart',
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ,
    UNIQUE(user_id, sentence_id, reaction_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_sentences_reactions;
-- +goose StatementEnd
