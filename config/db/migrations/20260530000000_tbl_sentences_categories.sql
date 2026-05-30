-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_sentences_categories;

CREATE TABLE tbl_sentences_categories (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    status_id BIGINT NOT NULL DEFAULT 1,
    "order" INTEGER NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    description TEXT,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_by BIGINT,
    deleted_at TIMESTAMPTZ
);

-- Seed initial categories
INSERT INTO tbl_sentences_categories (name, display_name) VALUES
('daily-life', 'Daily Life'),
('travel', 'Travel'),
('airport', 'Airport'),
('restaurant', 'Restaurant'),
('hospital', 'Hospital'),
('banking', 'Banking'),
('job-interview', 'Job Interview'),
('office', 'Office'),
('shopping', 'Shopping'),
('tech-support', 'Tech Support'),
('school', 'School'),
('sports', 'Sports'),
('phone-call', 'Phone Call'),
('emergency', 'Emergency'),
('renting', 'Renting'),
('general', 'General'),
('deep-sea-exploration', 'Deep Sea Exploration'),
('space-travel', 'Space Travel')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_sentences_categories;
-- +goose StatementEnd
