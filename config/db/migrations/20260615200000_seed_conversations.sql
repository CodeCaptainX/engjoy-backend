-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_conversations (uuid, text, title, source, source_type, source_detail, category, status_id, "order", created_by, created_at, updated_by, updated_at) 
VALUES (
    gen_random_uuid(), 
    'Hey, have you been studying Japanese lately? Yeah, just started last week, it is tough but fun!', 
    'Language Learning Chat', 
    'Daily Conversation', 
    'text', 
    'Casual chat between friends', 
    'Education', 
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
DELETE FROM tbl_conversations WHERE title = 'Language Learning Chat';
-- +goose StatementEnd
