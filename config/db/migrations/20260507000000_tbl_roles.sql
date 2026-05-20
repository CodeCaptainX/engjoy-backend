-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_roles (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tbl_roles_name ON tbl_roles(LOWER(name));
CREATE INDEX IF NOT EXISTS idx_tbl_roles_status_id ON tbl_roles(status_id);

INSERT INTO tbl_roles (
    id,
    name,
    description,
    is_admin,
    status_id
)
VALUES
(
    1,
    'admin',
    'Full system administrator',
    TRUE,
    1
),
(
    2,
    'user',
    'Default application user',
    FALSE,
    1
)
ON CONFLICT (id) DO NOTHING;

SELECT setval('tbl_roles_id_seq', (SELECT MAX(id) FROM tbl_roles));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_roles;
-- +goose StatementEnd
