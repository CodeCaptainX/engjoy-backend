package model

import "time"

type LearningFeedItem struct {
	ID        int64     `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Content   string    `db:"content" json:"content"`
	Type      string    `db:"type" json:"type"` // 'sentence' or 'conversation'
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	CreatedBy int64     `db:"created_by" json:"createdBy"`
}
