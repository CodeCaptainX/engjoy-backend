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

CREATE INDEX IF NOT EXISTS idx_tbl_users_email ON tbl_users(LOWER(email));
CREATE INDEX IF NOT EXISTS idx_tbl_users_deleted_at ON tbl_users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tbl_users_uuid ON tbl_users(uuid);
CREATE INDEX IF NOT EXISTS idx_tbl_users_status_id ON tbl_users(status_id);
