-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id BIGINT NOT NULL DEFAULT 2 REFERENCES tbl_roles(id),
    last_login_at TIMESTAMPTZ,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tbl_users_email ON tbl_users(LOWER(email));
CREATE INDEX IF NOT EXISTS idx_tbl_users_deleted_at ON tbl_users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_users_uuid ON tbl_users(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_users_status_id ON tbl_users(status_id);
CREATE INDEX IF NOT EXISTS idx_tbl_users_role_id ON tbl_users(role_id);

INSERT INTO tbl_users (
    name,
    email,
    password_hash,
    role_id,
    status_id
)
VALUES
(
    'Admin User',
    'admin@example.com',
    '$2a$10$hashedpassword1',
    1,
    1
),
(
    'John Doe',
    'john@example.com',
    '$2a$10$hashedpassword2',
    2,
    1
)
ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_users;
-- +goose StatementEnd
