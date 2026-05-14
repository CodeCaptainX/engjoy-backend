package model

import (
	"time"

	types "sentenceminer/pkg/model"
)

type Sentence struct {
	ID             int64      `db:"id" json:"id"`
	Text           string     `db:"text" json:"text"`
	Source         string     `db:"source" json:"source"`
	Category       string     `db:"category" json:"category"`
	ReviewCount    int        `db:"review_count" json:"review_count"`
	ReviewInterval int        `db:"review_interval" json:"review_interval"`
	EaseFactor     float64    `db:"ease_factor" json:"ease_factor"`
	LastRating     *string    `db:"last_rating" json:"last_rating,omitempty"`
	LastReviewedAt *time.Time `db:"last_reviewed_at" json:"last_reviewed_at,omitempty"`
	NextReviewAt   *time.Time `db:"next_review_at" json:"next_review_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	DeletedAt      *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
type SentenceResponse struct {
	Sentence Sentence `json:"sentence"`
}

type Analysis struct {
	ID           int64      `db:"id" json:"id"`
	SentenceID   int64      `db:"sentence_id" json:"sentence_id"`
	Explanation  string     `db:"explanation" json:"explanation"`
	Vocabulary   string     `db:"vocabulary" json:"vocabulary"`
	GrammarFocus string     `db:"grammar_focus" json:"grammar_focus"`
	Example      string     `db:"example" json:"example"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type SentenceWithAnalysis struct {
	SentenceID     int64      `db:"sentence_id" json:"sentence_id"`
	Text           string     `db:"text" json:"text"`
	Source         string     `db:"source" json:"source"`
	Category       string     `db:"category" json:"category"`
	ReviewCount    int        `db:"review_count" json:"review_count"`
	ReviewInterval int        `db:"review_interval" json:"review_interval"`
	EaseFactor     float64    `db:"ease_factor" json:"ease_factor"`
	LastRating     *string    `db:"last_rating" json:"last_rating,omitempty"`
	LastReviewedAt *time.Time `db:"last_reviewed_at" json:"last_reviewed_at,omitempty"`
	NextReviewAt   *time.Time `db:"next_review_at" json:"next_review_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	AnalysisID     *int64     `db:"analysis_id" json:"analysis_id"`
	Explanation    *string    `db:"explanation" json:"explanation"`
	Vocabulary     *string    `db:"vocabulary" json:"vocabulary"`
	GrammarFocus   *string    `db:"grammar_focus" json:"grammar_focus"`
	Example        *string    `db:"example" json:"example"`
	AnalyzedAt     *time.Time `db:"analyzed_at" json:"analyzed_at"`
}

type CreateSentenceRequest struct {
	Text     string `json:"text"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

type SentencesShowRequest struct {
	PagingOptions types.Paging   `json:"pagingOptions" query:"paging_options"`
	Filters       []types.Filter `json:"filters" query:"filters"`
	Sorts         []types.Sort   `json:"sorts" query:"sorts"`
	Offset        int            `json:"offset" query:"offset"`
}
type TTSRequest struct {
	Text string `json:"text"`
	Mode string `json:"mode"`
}

type ReviewRatingRequest struct {
	Rating string `json:"rating"`
}

type CategoryFeedRequest struct {
	Category   string  `json:"category"`
	ExcludeIDs []int64 `json:"excludeIds"`
	Limit      int     `json:"limit"`
	Focus      string  `json:"focus"`
}
