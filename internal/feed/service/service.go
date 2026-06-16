package service

import (
	"fmt"
	"sentenceminer/internal/feed/model"
	"sentenceminer/internal/feed/repository"

	"github.com/jmoiron/sqlx"
)

type FeedService struct {
	repo *repository.FeedRepository
}

func NewFeedService(db *sqlx.DB) *FeedService {
	return &FeedService{repo: repository.NewFeedRepository(db)}
}

func (s *FeedService) GetLearningFeed(userID int64, page, perPage int) ([]model.LearningFeedItem, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	
	items, err := s.repo.ListLearningFeed(userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get learning feed: %w", err)
	}
	return items, nil
}
