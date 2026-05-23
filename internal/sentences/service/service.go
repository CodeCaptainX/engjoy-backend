package service

import (
	"context"
	"fmt"
	"log"
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

func (s *SentenceService) Show(req postgres.QueryParamRequest) ([]model.SentenceWithAnalysis, int, *apiresponse.ErrorResponse) {
	return s.repo.Show(req)
}

func (s *SentenceService) List() ([]model.SentenceWithAnalysis, error) {
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

func (s *SentenceService) GenerateAndSaveSentences(ctx context.Context, category, focus string, count int) ([]model.Sentence, error) {
	if remaining := s.geminiCooldownRemaining(); remaining > 0 {
		return nil, fmt.Errorf("gemini is on cooldown, try again in %v", remaining)
	}

	normalizedCategory := strings.TrimSpace(strings.ToLower(category))
	if normalizedCategory == "" {
		normalizedCategory = "general"
	}

	existingTexts, err := s.repo.ListCategoryTexts(normalizedCategory, 100)
	if err != nil {
		return nil, err
	}

	generatedTexts, err := s.client.GenerateCategorySentences(ctx, normalizedCategory, focus, existingTexts, count)
	if err != nil {
		if geminiFailureReason(err) == "quota" {
			s.startGeminiCooldown(1 * time.Hour)
		}
		return nil, err
	}

	return s.repo.InsertGeneratedSentences(normalizedCategory, "ai-generated-flexible", generatedTexts)
}

func (s *SentenceService) StartDailyGenerationWorker() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			// Initial delay if needed or run immediately on first start
			// For now, let's just wait for the first tick
			<-ticker.C
			log.Println("[SentenceMiner] starting automatic daily sentence generation...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			_, err := s.GenerateAndSaveSentences(ctx, "general", "", 20)
			cancel()
			if err != nil {
				log.Printf("[SentenceMiner] daily generation failed: %v", err)
				// If quota error, we could retry in 1 hour
				if strings.Contains(strings.ToLower(err.Error()), "quota") {
					time.Sleep(1 * time.Hour)
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					_, _ = s.GenerateAndSaveSentences(ctx, "general", "", 20)
					cancel()
				}
			}
		}
	}()
}
