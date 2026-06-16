-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_users_conversations (uuid, title, category, user_id, is_public, status_id, "order", created_by, created_at, updated_by, updated_at) 
VALUES (
    gen_random_uuid(), 
    'Public Language Learning Chat', 
    'Education', 
    1, 
    TRUE, 
    1, 
    0, 
    1, 
    CURRENT_TIMESTAMP, 
    1, 
    CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_users_conversations WHERE title = 'Public Language Learning Chat';
-- +goose StatementEnd
