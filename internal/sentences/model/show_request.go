package model

import (
	"fmt"
	"strings"

	"sentenceminer/pkg/postgres"
)

func normalizeSortDirection(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "asc" {
		return "asc"
	}
	return "desc"
}

var sentenceShowColumns = map[string]bool{
	"id":               true,
	"sentence_id":      true,
	"text":             true,
	"source":           true,
	"category":         true,
	"review_count":     true,
	"review_interval":  true,
	"ease_factor":      true,
	"last_rating":      true,
	"last_reviewed_at": true,
	"next_review_at":   true,
	"created_at":       true,
}

func SanitizeSentenceShowRequest(req postgres.QueryParamRequest, alias string) (postgres.QueryParamRequest, error) {
	var err error
	req.Filters, err = sanitizeSentenceFilters(req.Filters, alias)
	if err != nil {
		return req, err
	}
	req.Sorts, err = sanitizeSentenceSorts(req.Sorts, alias)
	if err != nil {
		return req, err
	}
	return req, nil
}

func sanitizeSentenceFilters(filters []postgres.Filter, alias string) ([]postgres.Filter, error) {
	cleaned := make([]postgres.Filter, 0, len(filters))
	for _, filter := range filters {
		property, err := sanitizeSentenceProperty(filter.Property, alias)
		if err != nil {
			return nil, err
		}
		filter.Property = property
		cleaned = append(cleaned, filter)
	}
	return cleaned, nil
}

func sanitizeSentenceSorts(sorts []postgres.Sort, alias string) ([]postgres.Sort, error) {
	if len(sorts) == 0 {
		sorts = []postgres.Sort{{Property: "created_at", Direction: "desc"}}
	}

	cleaned := make([]postgres.Sort, 0, len(sorts))
	for _, sort := range sorts {
		property, err := sanitizeSentenceProperty(sort.Property, alias)
		if err != nil {
			return nil, err
		}
		sort.Property = property
		sort.Direction = normalizeSortDirection(sort.Direction)
		cleaned = append(cleaned, sort)
	}
	return cleaned, nil
}

func sanitizeSentenceProperty(property string, alias string) (string, error) {
	property = strings.TrimSpace(property)
	operator := ""
	if strings.Contains(property, "__") {
		parts := strings.SplitN(property, "__", 2)
		property = parts[0]
		operator = "__" + parts[1]
	}
	property = strings.TrimPrefix(property, "s.")
	if property == "sentence_id" {
		property = "id"
	}
	if !sentenceShowColumns[property] {
		return "", fmt.Errorf("invalid sentence query property: %s", property)
	}
	if alias != "" {
		property = alias + "." + property
	}
	return property + operator, nil
}
