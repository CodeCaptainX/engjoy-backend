package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sentenceminer/internal/sentences/model"
	"sentenceminer/pkg/postgres"
	"sentenceminer/pkg/sentencepack"

	"github.com/jmoiron/sqlx"
)

func (r *SentenceRepository) CreateSentence(text, source, category string) (model.Sentence, error) {
	var s model.Sentence
	normalizedCategory := normalizeCategory(category)
	err := r.db.QueryRowx(
		`INSERT INTO tbl_sentences (text, source, category) VALUES ($1, $2, $3)
		RETURNING id, text, source, category, review_count, review_interval, ease_factor,
		          last_rating, last_reviewed_at, next_review_at, created_at, deleted_at`,
		text,
		source,
		normalizedCategory,
	).StructScan(&s)
	return s, err
}

func (r *SentenceRepository) ImportEnvironmentPack() (int, error) {
	imported := 0
	for _, item := range sentencepack.EnvironmentSentencePack {
		var createdID int64
		err := r.db.QueryRowx(
			`INSERT INTO tbl_sentences (text, source, category)
			 SELECT $1, $2, $3
			 WHERE NOT EXISTS (
			 	SELECT 1
			 	FROM tbl_sentences
			 	WHERE LOWER(text) = LOWER($1) AND LOWER(category) = LOWER($3) AND deleted_at IS NULL
			 )
			 RETURNING id`,
			item.Text,
			item.Source,
			normalizeCategory(item.Category),
		).Scan(&createdID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return imported, err
		}
		imported++
	}

	return imported, nil
}

func (r *SentenceRepository) ImportStaticSentencePacks() (int, error) {
	imported := 0
	for _, category := range sentencepack.StaticSentencePackCategories() {
		entries := sentencepack.StaticCategoryEntries(category, "all", nil, 1000)
		inserted, err := r.InsertStaticSentenceEntries(category, "static-pack", entries)
		if err != nil {
			return imported, err
		}
		imported += len(inserted)
	}
	return imported, nil
}

func (r *SentenceRepository) ListCategoryTexts(category string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Queryx(
		`SELECT text
		FROM tbl_sentences
		WHERE deleted_at IS NULL AND category = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		normalizeCategory(category),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, limit)
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		items = append(items, text)
	}

	return items, rows.Err()
}

func (r *SentenceRepository) InsertGeneratedSentences(category, source string, texts []string) ([]model.Sentence, error) {
	normalizedCategory := normalizeCategory(category)
	inserted := make([]model.Sentence, 0, len(texts))
	seen := make(map[string]struct{}, len(texts))

	for _, text := range texts {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		var sentence model.Sentence
		err := r.db.QueryRowx(
			`INSERT INTO tbl_sentences (text, source, category)
			 SELECT $1, $2, $3
			 WHERE NOT EXISTS (
			 	SELECT 1
			 	FROM tbl_sentences
			 	WHERE LOWER(BTRIM(text)) = LOWER(BTRIM($1))
			 	  AND LOWER(category) = LOWER($3)
			 	  AND deleted_at IS NULL
			 )
			 RETURNING id, text, source, category, review_count, review_interval, ease_factor,
			           last_rating, last_reviewed_at, next_review_at, created_at, deleted_at`,
			trimmed,
			source,
			normalizedCategory,
		).StructScan(&sentence)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return inserted, err
		}

		inserted = append(inserted, sentence)
	}

	return inserted, nil
}

func (r *SentenceRepository) InsertStaticSentenceEntries(category, source string, entries []sentencepack.StaticSentenceEntry) ([]model.Sentence, error) {
	normalizedCategory := normalizeCategory(category)
	inserted := make([]model.Sentence, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry.Sentence)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		var sentence model.Sentence
		err := r.db.QueryRowx(
			`INSERT INTO tbl_sentences (text, source, category)
			 SELECT $1, $2, $3
			 WHERE NOT EXISTS (
			 	SELECT 1
			 	FROM tbl_sentences
			 	WHERE LOWER(BTRIM(text)) = LOWER(BTRIM($1))
			 	  AND LOWER(category) = LOWER($3)
			 	  AND deleted_at IS NULL
			 )
			 RETURNING id, text, source, category, review_count, review_interval, ease_factor,
			           last_rating, last_reviewed_at, next_review_at, created_at, deleted_at`,
			trimmed,
			source,
			normalizedCategory,
		).StructScan(&sentence)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return inserted, err
		}

		vocabJSON := "[]"
		if b, err := json.Marshal(entry.Vocabulary); err == nil {
			vocabJSON = string(b)
		}

		if _, err := r.CreateAnalysis(sentence.ID, entry.Meaning, vocabJSON, entry.GrammarFocus, entry.Example); err != nil {
			return inserted, err
		}

		inserted = append(inserted, sentence)
	}

	return inserted, nil
}

func (r *SentenceRepository) ListRandomByCategory(category string, excludeIDs []int64, limit int) ([]model.SentenceWithAnalysis, error) {
	if limit <= 0 {
		limit = 12
	}

	baseQuery := `
		SELECT s.id AS sentence_id, s.text, s.source, s.category, s.review_count, s.review_interval,
		       s.ease_factor, s.last_rating, s.last_reviewed_at, s.next_review_at, s.created_at,
		       a.id AS analysis_id, a.explanation, a.vocabulary, a.grammar_focus, a.example, a.created_at AS analyzed_at
		FROM tbl_sentences s
		LEFT JOIN LATERAL (
			SELECT id, explanation, vocabulary, grammar_focus, example, created_at
			FROM tbl_analyses
			WHERE sentence_id = s.id AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		) a ON true
		WHERE s.deleted_at IS NULL
	`

	args := []any{}
	if normalized := normalizeCategory(category); normalized != "all" {
		baseQuery += " AND s.category = ?"
		args = append(args, normalized)
	}
	if len(excludeIDs) > 0 {
		baseQuery += " AND s.id NOT IN (?)"
		args = append(args, excludeIDs)
	}
	baseQuery += " ORDER BY RANDOM() LIMIT ?"
	args = append(args, limit)

	query, queryArgs, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	items := []model.SentenceWithAnalysis{}
	if err := r.db.Select(&items, query, queryArgs...); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SentenceRepository) CreateAnalysis(sentenceID int64, explanation, vocabulary, grammarFocus, example string) (model.Analysis, error) {
	var a model.Analysis
	err := r.db.QueryRowx(
		`INSERT INTO tbl_analyses (sentence_id, explanation, vocabulary, grammar_focus, example)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sentence_id, explanation, vocabulary, grammar_focus, example, created_at, deleted_at`,
		sentenceID,
		explanation,
		vocabulary,
		grammarFocus,
		example,
	).StructScan(&a)
	return a, err
}

func (r *SentenceRepository) ListSentences(page int, limit int) ([]model.SentenceWithAnalysis, int, error) {
	total, err := r.CountSentences()
	if err != nil {
		return nil, 0, err
	}

	items := []model.SentenceWithAnalysis{}
	query := `
		SELECT s.id AS sentence_id, s.text, s.source, s.category, s.review_count, s.review_interval,
		       s.ease_factor, s.last_rating, s.last_reviewed_at, s.next_review_at, s.created_at,
		       a.id AS analysis_id, a.explanation, a.vocabulary, a.grammar_focus, a.example, a.created_at AS analyzed_at
		FROM tbl_sentences s
		LEFT JOIN LATERAL (
			SELECT id, explanation, vocabulary, grammar_focus, example, created_at
			FROM tbl_analyses
			WHERE sentence_id = s.id AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		) a ON true
		WHERE s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		` + postgres.BuildPaging(page, limit)

	err = r.db.Select(&items, query)
	return items, total, err
}

func (r *SentenceRepository) CountSentences() (int, error) {
	var total int
	err := r.db.Get(&total, `
		SELECT COUNT(*)
		FROM tbl_sentences
		WHERE deleted_at IS NULL
	`)
	return total, err
}

func (r *SentenceRepository) GetSentence(id int64) (model.Sentence, error) {
	var s model.Sentence
	err := r.db.QueryRowx(
		`SELECT id, text, source, category, review_count, review_interval, ease_factor,
		        last_rating, last_reviewed_at, next_review_at, created_at, deleted_at
		FROM tbl_sentences
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).StructScan(&s)
	return s, err
}

func (r *SentenceRepository) RateSentenceReview(id int64, rating string) (model.Sentence, error) {
	validRatings := map[string]bool{
		"again": true,
		"hard":  true,
		"good":  true,
		"easy":  true,
	}
	if !validRatings[rating] {
		return model.Sentence{}, fmt.Errorf("invalid rating: %s", rating)
	}

	current, err := r.GetSentence(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Sentence{}, ErrSentenceNotFound
		}
		return model.Sentence{}, err
	}

	now := time.Now().UTC()
	nextReviewAt := now
	nextInterval := current.ReviewInterval
	nextEase := current.EaseFactor

	switch rating {
	case "again":
		nextInterval = 0
		nextEase = maxFloat(1.3, current.EaseFactor-0.2)
		nextReviewAt = now.Add(6 * time.Hour)
	case "hard":
		nextInterval = maxInt(1, current.ReviewInterval)
		nextEase = maxFloat(1.4, current.EaseFactor-0.05)
		nextReviewAt = now.Add(24 * time.Hour)
	case "good":
		if current.ReviewCount == 0 {
			nextInterval = 1
		} else {
			nextInterval = maxInt(2, int(float64(maxInt(1, current.ReviewInterval))*current.EaseFactor))
		}
		nextReviewAt = now.Add(time.Duration(nextInterval) * 24 * time.Hour)
	case "easy":
		baseInterval := maxInt(1, current.ReviewInterval)
		if current.ReviewCount == 0 {
			baseInterval = 3
		}
		nextInterval = maxInt(3, int(float64(baseInterval)*(current.EaseFactor+0.35)))
		nextEase = current.EaseFactor + 0.15
		nextReviewAt = now.Add(time.Duration(nextInterval) * 24 * time.Hour)
	}

	var updated model.Sentence
	err = r.db.QueryRowx(
		`UPDATE tbl_sentences
		 SET review_count = review_count + 1,
		     review_interval = $2,
		     ease_factor = $3,
		     last_rating = $4,
		     last_reviewed_at = $5,
		     next_review_at = $6,
		     updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING id, text, source, category, review_count, review_interval, ease_factor,
		           last_rating, last_reviewed_at, next_review_at, created_at, deleted_at`,
		id,
		nextInterval,
		nextEase,
		rating,
		now,
		nextReviewAt,
	).StructScan(&updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Sentence{}, ErrSentenceNotFound
		}
		return model.Sentence{}, err
	}

	return updated, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func normalizeCategory(category string) string {
	value := strings.TrimSpace(strings.ToLower(category))
	if value == "" {
		return "general"
	}
	return value
}

func (r *SentenceRepository) SoftDeleteSentence(id int64) (time.Time, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return time.Time{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var deletedAt time.Time
	row := tx.QueryRowx(
		`UPDATE tbl_sentences
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING deleted_at`,
		id,
	)
	if scanErr := row.Scan(&deletedAt); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			err = ErrSentenceNotFound
			return time.Time{}, err
		}
		err = scanErr
		return time.Time{}, err
	}

	if _, execErr := tx.Exec(
		`UPDATE tbl_analyses
		SET deleted_at = $2
		WHERE sentence_id = $1 AND deleted_at IS NULL`,
		id,
		deletedAt,
	); execErr != nil {
		err = execErr
		return time.Time{}, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		err = commitErr
		return time.Time{}, err
	}

	return deletedAt, nil
}
