-- +goose Up
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name='tbl_users_conversations' AND column_name='is_public'
    ) THEN
        ALTER TABLE tbl_users_conversations ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
END
$$;

-- +goose Down
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name='tbl_users_conversations' AND column_name='is_public'
    ) THEN
        ALTER TABLE tbl_users_conversations DROP COLUMN is_public;
    END IF;
END
$$;
