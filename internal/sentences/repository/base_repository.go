package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sentenceminer/internal/sentences/model"
	apiresponse "sentenceminer/pkg/http/response"
	customlog "sentenceminer/pkg/logs"
	"sentenceminer/pkg/postgres"

	"github.com/jmoiron/sqlx"
)

var ErrSentenceNotFound = errors.New("sentence not found")

type SentenceRepository struct {
	db *sqlx.DB
}

func NewSentenceRepository(db *sqlx.DB) *SentenceRepository {
	return &SentenceRepository{db: db}
}

func (r *SentenceRepository) Create(text, source, category string) (*model.SentenceResponse, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var sentence model.Sentence
	if strings.TrimSpace(source) == "" {
		source = "api"
	}
	if strings.TrimSpace(category) == "" {
		category = "general"
	}

	err = tx.QueryRowx(
		`INSERT INTO tbl_sentences (text, source, category)
		VALUES ($1, $2, $3)
		RETURNING id, text, source, category, review_count, review_interval, ease_factor,
		          last_rating, last_reviewed_at, next_review_at, created_at, deleted_at`,
		strings.TrimSpace(text),
		strings.TrimSpace(source),
		strings.ToLower(strings.TrimSpace(category)),
	).StructScan(&sentence)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &model.SentenceResponse{Sentence: sentence}, nil
}

func (r *SentenceRepository) Show(req postgres.QueryParamRequest) ([]model.Sentence, int, *apiresponse.ErrorResponse) {
	req, err := model.SanitizeSentenceShowRequest(req, "")
	if err != nil {
		return nil, 0, customlog.NewCustomLogResponse(customlog.ErrorParams{
			LogMessage: "sentence_show_invalid_query",
			I18nKey:    "sentence_show_invalid_query",
			Err:        err,
			TypeError:  "error",
			StatusCode: 400,
		})
	}

	filterSQL, args := postgres.BuildSQLFilter(req.Filters)
	total, err := r.Count(filterSQL, args)
	if err != nil {
		return nil, 0, customlog.NewCustomLogResponse(customlog.ErrorParams{
			LogMessage: "sentence_show_count_failed",
			I18nKey:    "sentence_show_count_failed",
			Err:        err,
			TypeError:  "error",
			StatusCode: 500,
		})
	}

	whereSQL := "deleted_at IS NULL"
	if strings.TrimSpace(filterSQL) != "" {
		whereSQL += " AND " + filterSQL
	}

	items := []model.Sentence{}
	query := fmt.Sprintf(`
		SELECT id, text, source, category, review_count, review_interval, ease_factor,
		       last_rating, last_reviewed_at, next_review_at, created_at, deleted_at
		FROM tbl_sentences
		WHERE %s
		%s
		%s
	`, whereSQL, postgres.BuildSQLSort(req.Sorts), postgres.BuildPaging(req.PagingOptions.Page, req.PagingOptions.PerPage))

	err = r.db.Select(&items, query, args...)
	if err != nil {
		return nil, 0, customlog.NewCustomLogResponse(customlog.ErrorParams{
			LogMessage: "sentence_show_fetch_failed",
			I18nKey:    "sentence_show_fetch_failed",
			Err:        err,
			TypeError:  "error",
			StatusCode: 500,
		})
	}

	return items, total, nil
}

func (r *SentenceRepository) Count(filterSQL string, args []interface{}) (int, error) {
	var total int
	whereSQL := "deleted_at IS NULL"
	if strings.TrimSpace(filterSQL) != "" {
		whereSQL += " AND " + filterSQL
	}
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tbl_sentences
		WHERE %s
	`, whereSQL)
	err := r.db.Get(&total, query, args...)
	return total, err
}

func (r *SentenceRepository) Get(id int64) (model.Sentence, error) {
	var sentence model.Sentence
	err := r.db.QueryRowx(
		`SELECT id, text, source, category, review_count, review_interval, ease_factor,
		        last_rating, last_reviewed_at, next_review_at, created_at, deleted_at
		FROM tbl_sentences
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).StructScan(&sentence)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Sentence{}, ErrSentenceNotFound
		}
		return model.Sentence{}, err
	}
	return sentence, nil
}

func (r *SentenceRepository) SoftDelete(id int64) (time.Time, error) {
	var deletedAt time.Time
	err := r.db.QueryRowx(
		`UPDATE tbl_sentences
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING deleted_at`,
		id,
	).Scan(&deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ErrSentenceNotFound
		}
		return time.Time{}, err
	}
	return deletedAt, nil
}
