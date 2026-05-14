package service

import (
	"strings"
	"sync"
	"time"

	config "sentenceminer/config"
	"sentenceminer/internal/sentences/model"
	"sentenceminer/internal/sentences/repository"
	apiresponse "sentenceminer/pkg/http/response"
	"sentenceminer/pkg/postgres"

	"github.com/jmoiron/sqlx"
)

type SentenceService struct {
	repo                *repository.SentenceRepository
	client              *Client
	geminiCooldownMu    sync.Mutex
	geminiCooldownUntil time.Time
}

func NewSentenceService(db *sqlx.DB) *SentenceService {
	cfg := config.NewConfig()
	return &SentenceService{
		repo:   repository.NewSentenceRepository(db),
		client: NewClient(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiTTSModel, cfg.GeminiBase),
	}
}

func (s *SentenceService) Create(req model.CreateSentenceRequest) (*model.SentenceResponse, error) {
	return s.repo.Create(strings.TrimSpace(req.Text), req.Source, req.Category)
}

func (s *SentenceService) Show(req postgres.QueryParamRequest) ([]model.Sentence, int, *apiresponse.ErrorResponse) {
	return s.repo.Show(req)
}

func (s *SentenceService) List() ([]model.Sentence, error) {
	items, _, err := s.repo.Show(postgres.QueryParamRequest{})
	if err != nil {
		return nil, err.Err
	}
	return items, nil
}

func (s *SentenceService) Get(id int64) (model.Sentence, error) {
	return s.repo.Get(id)
}

func (s *SentenceService) ImportEnvironmentPack() (int, error) {
	return s.repo.ImportEnvironmentPack()
}

func (s *SentenceService) ListSentences(page, limit int) ([]model.SentenceWithAnalysis, int, error) {
	return s.repo.ListSentences(page, limit)
}

func (s *SentenceService) GetSentence(id int64) (model.Sentence, error) {
	return s.repo.GetSentence(id)
}

func (s *SentenceService) RateSentenceReview(id int64, rating string) (model.Sentence, error) {
	return s.repo.RateSentenceReview(id, rating)
}

func (s *SentenceService) SoftDeleteSentence(id int64) (time.Time, error) {
	return s.repo.SoftDeleteSentence(id)
}
