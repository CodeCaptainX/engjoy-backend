-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_users_conversations (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    title VARCHAR(255) NULL,
    source VARCHAR(255) NULL,
    category VARCHAR(50) NULL,
    user_id BIGINT NOT NULL,
    status_id INTEGER NOT NULL DEFAULT 1,
    "order" INTEGER NOT NULL DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,

    CONSTRAINT fk_user_owner
        FOREIGN KEY (user_id)
        REFERENCES tbl_users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_created_by
        FOREIGN KEY (created_by)
        REFERENCES tbl_users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES tbl_users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_deleted_by
        FOREIGN KEY (deleted_by)
        REFERENCES tbl_users(id)
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tbl_users_conversations_messages (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    conversation_id BIGINT NOT NULL,
    speaker VARCHAR(255) NOT NULL,
    message_text TEXT NOT NULL,
    message_order INTEGER NOT NULL,
    status_id INTEGER NOT NULL DEFAULT 1,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,

    CONSTRAINT fk_message_conversation
        FOREIGN KEY (conversation_id)
        REFERENCES tbl_users_conversations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_message_created_by
        FOREIGN KEY (created_by)
        REFERENCES tbl_users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_message_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES tbl_users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_message_deleted_by
        FOREIGN KEY (deleted_by)
        REFERENCES tbl_users(id)
        ON DELETE SET NULL
);

-- +goose Down
DROP TABLE IF EXISTS tbl_users_conversations_messages;
DROP TABLE IF EXISTS tbl_users_conversations;
