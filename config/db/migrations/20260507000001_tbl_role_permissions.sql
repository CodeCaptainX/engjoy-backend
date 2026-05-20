-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_role_permissions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    role_id BIGINT NOT NULL REFERENCES tbl_roles(id) ON DELETE CASCADE,
    permission_key TEXT NOT NULL,
    can_view BOOLEAN NOT NULL DEFAULT FALSE,
    can_create BOOLEAN NOT NULL DEFAULT FALSE,
    can_update BOOLEAN NOT NULL DEFAULT FALSE,
    can_delete BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ,
    UNIQUE (role_id, permission_key)
);

CREATE INDEX IF NOT EXISTS idx_tbl_role_permissions_role_id ON tbl_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_tbl_role_permissions_permission_key ON tbl_role_permissions(permission_key);

INSERT INTO tbl_role_permissions (
    role_id,
    permission_key,
    can_view,
    can_create,
    can_update,
    can_delete,
    can_manage,
    status_id
)
VALUES
(
    1,
    'admin',
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    1
),
(
    2,
    'sentences',
    TRUE,
    TRUE,
    TRUE,
    FALSE,
    FALSE,
    1
)
ON CONFLICT (role_id, permission_key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_role_permissions;
-- +goose StatementEnd
