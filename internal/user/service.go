package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TouchLastLogin(id int64, at time.Time) error {
	return s.repo.TouchLastLogin(id, at)
}

func (s *Service) AddFavorite(ctx context.Context, userID, sentenceID int64) error {
	return s.repo.AddFavorite(userID, sentenceID)
}

func (s *Service) RemoveFavorite(ctx context.Context, userID, sentenceID int64) error {
	return s.repo.RemoveFavorite(userID, sentenceID)
}

func (s *Service) GetFavorites(ctx context.Context, userID int64) ([]int64, error) {
	return s.repo.ListFavorites(userID)
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
