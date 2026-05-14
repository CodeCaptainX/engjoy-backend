package sentencepack

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"strings"
)

//go:embed sentence_packs/*.json
var staticSentencePackFS embed.FS

type StaticVocabEntry struct {
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
}

type StaticSentenceEntry struct {
	Sentence     string             `json:"sentence"`
	Meaning      string             `json:"meaning"`
	GrammarFocus string             `json:"grammar_focus"`
	Vocabulary   []StaticVocabEntry `json:"vocabulary"`
	Example      string             `json:"example"`
}

type staticSentencePackFile struct {
	Category string                           `json:"category"`
	Focuses  map[string][]StaticSentenceEntry `json:"focuses"`
}

type staticSentencePack map[string]map[string][]StaticSentenceEntry

var cachedStaticSentencePack staticSentencePack

func StaticCategoryEntries(category, focus string, existing []string, limit int) []StaticSentenceEntry {
	if limit <= 0 {
		return nil
	}

	pack := loadStaticSentencePack()
	category = normalizePackKey(category)
	focus = normalizePackKey(focus)
	if focus == "" {
		focus = "all"
	}

	focusGroups, ok := pack[category]
	if !ok {
		focusGroups = pack["general"]
	}

	candidates := make([]StaticSentenceEntry, 0, limit)
	if focus != "all" {
		candidates = append(candidates, focusGroups[focus]...)
	}
	for group, entries := range focusGroups {
		if focus != "all" && group == focus {
			continue
		}
		candidates = append(candidates, entries...)
	}

	seen := make(map[string]struct{}, len(existing))
	for _, text := range existing {
		seen[normalizeSentenceKey(text)] = struct{}{}
	}

	results := make([]StaticSentenceEntry, 0, limit)
	for _, entry := range candidates {
		key := normalizeSentenceKey(entry.Sentence)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		entry.Sentence = strings.TrimSpace(entry.Sentence)
		results = append(results, entry)
		if len(results) >= limit {
			break
		}
	}

	return results
}

func StaticSentencePackCategories() []string {
	pack := loadStaticSentencePack()
	categories := make([]string, 0, len(pack))
	for category := range pack {
		categories = append(categories, category)
	}
	return categories
}

func loadStaticSentencePack() staticSentencePack {
	if cachedStaticSentencePack != nil {
		return cachedStaticSentencePack
	}

	files, err := staticSentencePackFS.ReadDir("sentence_packs")
	if err != nil {
		cachedStaticSentencePack = staticSentencePack{}
		return cachedStaticSentencePack
	}

	pack := staticSentencePack{}
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		content, err := staticSentencePackFS.ReadFile("sentence_packs/" + file.Name())
		if err != nil {
			continue
		}

		var parsed staticSentencePackFile
		if err := json.Unmarshal(content, &parsed); err != nil {
			continue
		}

		category := normalizePackKey(parsed.Category)
		if category == "" {
			category = normalizePackKey(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		}
		pack[category] = parsed.Focuses
	}

	cachedStaticSentencePack = pack
	return cachedStaticSentencePack
}

func normalizePackKey(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeSentenceKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
