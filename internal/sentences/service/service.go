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

func (s *SentenceService) CreateSentence(req model.CreateSentenceRequest, userID int64) (model.Sentence, error) {
	return s.repo.CreateSentence(strings.TrimSpace(req.Text), req.Source, req.Category, req.Explanation, userID)
}

func (s *SentenceService) Show(req postgres.QueryParamRequest, userID int64) ([]model.SentenceWithAnalysis, int, *apiresponse.ErrorResponse) {
	return s.repo.Show(req, userID)
}

func (s *SentenceService) List(userID int64) ([]model.SentenceWithAnalysis, error) {
	items, _, err := s.repo.Show(postgres.QueryParamRequest{}, userID)
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

func (s *SentenceService) ListSentences(page, limit int, userID int64) ([]model.SentenceWithAnalysis, int, error) {
	return s.repo.ListSentences(page, limit, userID)
}

func (s *SentenceService) GetSentence(id int64) (model.Sentence, error) {
	return s.repo.GetSentence(id)
}

func (s *SentenceService) RateSentenceReview(id int64, rating string) (model.Sentence, error) {
	return s.repo.RateSentenceReview(id, rating)
}

func (s *SentenceService) SoftDeleteSentence(id int64) (time.Time, error) {
	return s.repo.SoftDelete(id)
}

func (s *SentenceService) AddFavorite(userID int64, sentenceUUID string) error {
	return s.repo.AddFavorite(userID, sentenceUUID)
}

func (s *SentenceService) RemoveFavorite(userID int64, sentenceUUID string) error {
	return s.repo.RemoveFavorite(userID, sentenceUUID)
}

func (s *SentenceService) ListFavorites(userID int64, req postgres.QueryParamRequest) ([]model.SentenceWithAnalysis, int, error) {
	return s.repo.ListFavorites(userID, req)
}

func (s *SentenceService) ToggleReaction(userID int64, sentenceUUID string, reactionType string) (string, error) {
	return s.repo.ToggleReaction(userID, sentenceUUID, reactionType)
}

func (s *SentenceService) GetReactionCount(sentenceUUID string, reactionType string) (int, error) {
	return s.repo.GetReactionCount(sentenceUUID, reactionType)
}

func (s *SentenceService) GetRandomCategoryName() (string, error) {
	return s.repo.GetRandomCategoryName()
}

func (s *SentenceService) ListCategories() ([]model.SentenceCategory, error) {
	return s.repo.ListCategories()
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
		reason := geminiFailureReason(err)
		if reason == "daily_quota" {
			s.startGeminiCooldown(24 * time.Hour)
		} else if reason == "quota" {
			s.startGeminiCooldown(1 * time.Hour)
		}
		return nil, err
	}

	return s.repo.InsertGeneratedSentences(normalizedCategory, "ai-generated-flexible", generatedTexts)
}

func (s *SentenceService) StartDailyGenerationWorker() {
	// Trigger once on startup after a 5s delay
	go func() {
		time.Sleep(5 * time.Second)
		s.RunDailyGeneration()
	}()

	// Run every 1 hour to stay within the 20-request daily limit
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			<-ticker.C
			s.RunDailyGeneration()
		}
	}()
}

func (s *SentenceService) RunDailyGeneration() {
	// Pick exactly ONE random category from the database
	cat, err := s.repo.GetRandomCategoryName()
	if err != nil {
		log.Printf("[SentenceMiner] periodic generation: failed to fetch random category: %v", err)
		return
	}

	// Models to try in order (including the primary one from config)
	cfg := config.NewConfig()
	fallbackModels := []string{
		cfg.GeminiModel, // Start with the primary model (2.5-flash)
		"gemini-2.0-flash",
		"gemini-1.5-flash",
		"gemini-1.5-flash-8b",
	}

	log.Printf("[SentenceMiner] periodic generation: picked random category \"%s\" from database", cat)
	
	currentModelIdx := 0
	for currentModelIdx < len(fallbackModels) {
		// Set the model for this attempt
		s.client.SetModel(fallbackModels[currentModelIdx])

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		_, err := s.GenerateAndSaveSentences(ctx, cat, "useful expressions", 10)
		cancel()

		if err == nil {
			log.Printf("[SentenceMiner] successfully generated sentences for %s using %s", cat, fallbackModels[currentModelIdx])
			return // Success, exit this cycle
		}

		reason := geminiFailureReason(err)
		log.Printf("[SentenceMiner] generation failed for category %s with model %s: %v (reason: %s)", 
			cat, fallbackModels[currentModelIdx], err, reason)

		if reason == "daily_quota" {
			log.Printf("[SentenceMiner] daily quota reached for %s, switching to next model...", fallbackModels[currentModelIdx])
			currentModelIdx++
			if currentModelIdx < len(fallbackModels) {
				// We'll set the new model in the next iteration of the loop
				continue 
			} else {
				log.Println("[SentenceMiner] ALL models exhausted for today. Waiting for tomorrow.")
				s.startGeminiCooldown(24 * time.Hour)
				return
			}
		}

		if reason == "quota" {
			log.Println("[SentenceMiner] temporary rate limit reached, will retry in next cycle")
			return
		}

		// For other errors, just stop this cycle
		break
	}
}
