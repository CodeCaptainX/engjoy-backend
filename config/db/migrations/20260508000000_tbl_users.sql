-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    "order" INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    google_id TEXT UNIQUE,
    login_session TEXT,
    role_id BIGINT NOT NULL DEFAULT 2 REFERENCES tbl_roles(id),
    last_login_at TIMESTAMPTZ,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

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
