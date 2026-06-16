package repository

import (
	"fmt"
	"sentenceminer/internal/feed/model"

	"github.com/jmoiron/sqlx"
)

type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) ListLearningFeed(userID int64, limit, offset int) ([]model.LearningFeedItem, error) {
	query := `
		(
			SELECT id, uuid, text AS content, 'sentence' AS type, created_at, created_by
			FROM tbl_sentences
			WHERE deleted_at IS NULL
		)
		UNION ALL
		(
			SELECT id, uuid, title AS content, 'conversation' AS type, created_at, created_by
			FROM tbl_users_conversations
			WHERE deleted_at IS NULL AND (is_public = TRUE OR user_id = :user_id)
		)
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`
	
	args := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	nstmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for feed: %w", err)
	}
	defer nstmt.Close()

	items := []model.LearningFeedItem{}
	err = nstmt.Select(&items, args)
	if err != nil {
		return nil, fmt.Errorf("failed to select learning feed: %w", err)
	}

	return items, nil
}
