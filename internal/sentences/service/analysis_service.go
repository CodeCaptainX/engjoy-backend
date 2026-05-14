package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"sentenceminer/internal/sentences/model"
	"sentenceminer/pkg/sentencepack"
)

type GeminiResult struct {
	Explanation  string        `json:"explanation"`
	Vocabulary   []GeminiVocab `json:"vocabulary"`
	GrammarFocus string        `json:"grammar_focus"`
	Example      string        `json:"example"`
}

type GeminiVocab struct {
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
}

func (s *SentenceService) geminiCooldownRemaining() time.Duration {
	s.geminiCooldownMu.Lock()
	defer s.geminiCooldownMu.Unlock()

	remaining := time.Until(s.geminiCooldownUntil)
	if remaining <= 0 {
		return 0
	}
	return remaining
}

func (s *SentenceService) startGeminiCooldown(duration time.Duration) {
	s.geminiCooldownMu.Lock()
	defer s.geminiCooldownMu.Unlock()

	until := time.Now().Add(duration)
	if until.After(s.geminiCooldownUntil) {
		s.geminiCooldownUntil = until
	}
}

func logGeminiFailure(operation string, err error, fields map[string]any) {
	if err == nil {
		return
	}

	parts := []string{
		"[SentenceMiner] gemini_failed",
		"operation=" + operation,
		"reason=" + geminiFailureReason(err),
	}
	for key, value := range fields {
		parts = append(parts, key+"="+cleanLogValue(toLogValue(value)))
	}
	parts = append(parts, "error="+sanitizeGeminiError(err.Error()))
	log.Println(strings.Join(parts, " "))
}

func logGeminiSkipped(operation, reason string, fields map[string]any) {
	parts := []string{
		"[SentenceMiner] gemini_skipped",
		"operation=" + operation,
		"reason=" + reason,
	}
	for key, value := range fields {
		parts = append(parts, key+"="+cleanLogValue(toLogValue(value)))
	}
	log.Println(strings.Join(parts, " "))
}

func geminiFailureReason(err error) string {
	if err == nil {
		return "unknown"
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "api key is not configured"):
		return "missing_key"
	case strings.Contains(message, "status 401"), strings.Contains(message, "status 403"), strings.Contains(message, "permission_denied"), strings.Contains(message, "api key"):
		return "auth"
	case strings.Contains(message, "status 429"), strings.Contains(message, "quota"), strings.Contains(message, "rate"):
		return "quota"
	case strings.Contains(message, "context deadline"), strings.Contains(message, "timeout"), strings.Contains(message, "connection"), strings.Contains(message, "no such host"):
		return "network"
	case strings.Contains(message, "empty response"):
		return "empty_response"
	default:
		return "unknown"
	}
}

func toLogValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return fmt.Sprint(typed)
	}
}

func cleanLogValue(value string) string {
	cleaned := strings.ReplaceAll(value, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return strings.TrimSpace(cleaned)
}

func sanitizeGeminiError(message string) string {
	return cleanLogValue(message)
}

func (s *SentenceService) CreateAndAnalyze(ctx context.Context, text, source string) (model.Sentence, model.Analysis, error) {
	sentence, err := s.repo.CreateSentence(text, source, "general")
	if err != nil {
		return model.Sentence{}, model.Analysis{}, err
	}

	raw, err := s.client.AnalyzeSentence(ctx, text)
	if err != nil {
		logGeminiFailure("create_and_analyze", err, map[string]any{
			"source": source,
		})
		return sentence, model.Analysis{}, err
	}

	var parsed GeminiResult
	explanation := ""
	example := ""
	grammarFocus := ""
	vocabJSON := "[]"

	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		explanation = parsed.Explanation
		example = parsed.Example
		grammarFocus = parsed.GrammarFocus
		if b, err := json.Marshal(parsed.Vocabulary); err == nil {
			vocabJSON = string(b)
		}
	} else {
		explanation = raw
	}

	analysis, err := s.repo.CreateAnalysis(sentence.ID, explanation, vocabJSON, grammarFocus, example)
	if err != nil {
		return sentence, model.Analysis{}, err
	}

	return sentence, analysis, nil
}
func (s *SentenceService) AnalyzeExisting(ctx context.Context, sentenceID int64) (model.Sentence, model.Analysis, error) {
	// Get sentence from DB
	sentence, err := s.repo.GetSentence(sentenceID)
	if err != nil {
		return model.Sentence{}, model.Analysis{}, err
	}

	// Call Gemini
	raw, err := s.client.AnalyzeSentence(ctx, sentence.Text)
	if err != nil {
		logGeminiFailure("analyze_existing", err, map[string]any{
			"sentence_id": sentenceID,
			"category":    sentence.Category,
		})
		return sentence, model.Analysis{}, err
	}

	var parsed GeminiResult
	explanation := ""
	example := ""
	grammarFocus := ""
	vocabJSON := "[]"

	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		explanation = parsed.Explanation
		example = parsed.Example
		grammarFocus = parsed.GrammarFocus
		if b, err := json.Marshal(parsed.Vocabulary); err == nil {
			vocabJSON = string(b)
		}
	} else {
		explanation = raw
	}

	analysis, err := s.repo.CreateAnalysis(sentenceID, explanation, vocabJSON, grammarFocus, example)
	if err != nil {
		return sentence, model.Analysis{}, err
	}

	return sentence, analysis, nil
}

func (s *SentenceService) GenerateSpeech(ctx context.Context, text, mode string) ([]byte, error) {
	return s.client.GenerateSpeech(ctx, text, mode)
}

func (s *SentenceService) GetCategoryFeed(ctx context.Context, category, focus string, excludeIDs []int64, limit int) ([]model.SentenceWithAnalysis, int, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	normalizedCategory := strings.TrimSpace(strings.ToLower(category))
	if normalizedCategory == "" {
		normalizedCategory = "general"
	}
	normalizedFocus := strings.TrimSpace(strings.ToLower(focus))
	if normalizedFocus == "" {
		normalizedFocus = "all"
	}

	items, err := s.repo.ListRandomByCategory(normalizedCategory, excludeIDs, limit)
	if err != nil {
		return nil, 0, err
	}

	if len(items) >= limit || normalizedCategory == "all" {
		return items, 0, nil
	}

	existingTexts, err := s.repo.ListCategoryTexts(normalizedCategory, 120)
	if err != nil {
		return items, 0, err
	}

	requestedCount := limit - len(items) + 4
	staticEntries := sentencepack.StaticCategoryEntries(normalizedCategory, normalizedFocus, existingTexts, requestedCount)
	if len(staticEntries) > 0 {
		inserted, err := s.repo.InsertStaticSentenceEntries(normalizedCategory, "static-pack", staticEntries)
		if err != nil {
			return items, 0, err
		}
		return s.appendGeneratedCategoryItems(ctx, items, excludeIDs, normalizedCategory, limit, len(inserted))
	}

	if remaining := s.geminiCooldownRemaining(); remaining > 0 {
		logGeminiSkipped("category_feed", "cooldown", map[string]any{
			"category":       normalizedCategory,
			"focus":          normalizedFocus,
			"remaining_secs": int(remaining.Seconds()) + 1,
		})
		return items, 0, nil
	}

	generatedTexts, err := s.client.GenerateCategorySentences(ctx, normalizedCategory, normalizedFocus, existingTexts, requestedCount)
	if err != nil {
		logGeminiFailure("category_feed", err, map[string]any{
			"category":       normalizedCategory,
			"focus":          normalizedFocus,
			"existing_count": len(existingTexts),
			"requested":      requestedCount,
		})
		if geminiFailureReason(err) == "quota" {
			s.startGeminiCooldown(30 * time.Second)
		}
		return items, 0, nil
	}

	inserted, err := s.repo.InsertGeneratedSentences(normalizedCategory, "ai-generated", generatedTexts)
	if err != nil {
		return items, 0, err
	}

	return s.appendGeneratedCategoryItems(ctx, items, excludeIDs, normalizedCategory, limit, len(inserted))
}

func (s *SentenceService) appendGeneratedCategoryItems(ctx context.Context, items []model.SentenceWithAnalysis, excludeIDs []int64, category string, limit int, generated int) ([]model.SentenceWithAnalysis, int, error) {
	if len(items) >= limit {
		return items, generated, nil
	}
	allExclude := append([]int64{}, excludeIDs...)
	for _, item := range items {
		allExclude = append(allExclude, item.SentenceID)
	}

	moreItems, err := s.repo.ListRandomByCategory(category, allExclude, limit-len(items))
	if err != nil {
		return items, generated, err
	}

	items = append(items, moreItems...)
	return items, generated, nil
}
