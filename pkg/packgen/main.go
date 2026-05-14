package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	config "sentenceminer/config"
	"sentenceminer/internal/sentences/service"
	"sentenceminer/pkg/sentencepack"
)

type packFile struct {
	Category string                                        `json:"category"`
	Focuses  map[string][]sentencepack.StaticSentenceEntry `json:"focuses"`
}

func main() {
	category := flag.String("category", "", "sentence category, for example airport")
	focus := flag.String("focus", "all", "category focus, for example check-in")
	count := flag.Int("count", 20, "number of entries to generate, max 50")
	outDir := flag.String("out", filepath.Join("pkg", "sentencepack", "sentence_packs"), "output sentence pack directory")
	flag.Parse()

	normalizedCategory := normalizeKey(*category)
	normalizedFocus := normalizeKey(*focus)
	if normalizedCategory == "" {
		log.Fatal("--category is required")
	}
	if normalizedFocus == "" {
		normalizedFocus = "all"
	}

	packPath := filepath.Join(*outDir, normalizedCategory+".json")
	existingPack, err := readPack(packPath, normalizedCategory)
	if err != nil {
		log.Fatalf("read pack: %v", err)
	}

	existingSentences := collectSentences(existingPack)

	cfg := config.NewConfig()
	client := service.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiTTSModel, cfg.GeminiBase)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	raw, err := client.GenerateSentencePack(ctx, normalizedCategory, normalizedFocus, existingSentences, *count)
	if err != nil {
		log.Fatalf("generate sentence pack: %v", err)
	}

	var generated packFile
	if err := json.Unmarshal([]byte(raw), &generated); err != nil {
		fmt.Fprintln(os.Stderr, raw)
		log.Fatalf("parse generated JSON: %v", err)
	}

	if generated.Focuses == nil {
		log.Fatal("generated JSON missing focuses")
	}

	added := mergePack(existingPack, generated, normalizedCategory, normalizedFocus)
	if err := writePack(packPath, existingPack); err != nil {
		log.Fatalf("write pack: %v", err)
	}

	log.Printf("pack generated category=%s focus=%s added=%d path=%s", normalizedCategory, normalizedFocus, added, packPath)
}

func readPack(path, category string) (packFile, error) {
	pack := packFile{
		Category: category,
		Focuses:  map[string][]sentencepack.StaticSentenceEntry{},
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return pack, nil
		}
		return pack, err
	}

	if err := json.Unmarshal(content, &pack); err != nil {
		return pack, err
	}
	if pack.Category == "" {
		pack.Category = category
	}
	if pack.Focuses == nil {
		pack.Focuses = map[string][]sentencepack.StaticSentenceEntry{}
	}
	return pack, nil
}

func writePack(path string, pack packFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func mergePack(target packFile, generated packFile, category, focus string) int {
	if target.Category == "" {
		target.Category = category
	}
	if target.Focuses == nil {
		target.Focuses = map[string][]sentencepack.StaticSentenceEntry{}
	}

	seen := make(map[string]struct{})
	for _, entry := range collectSentences(target) {
		seen[normalizeSentence(entry)] = struct{}{}
	}

	added := 0
	for generatedFocus, entries := range generated.Focuses {
		targetFocus := normalizeKey(generatedFocus)
		if targetFocus == "" {
			targetFocus = focus
		}
		for _, entry := range entries {
			entry.Sentence = strings.TrimSpace(entry.Sentence)
			key := normalizeSentence(entry.Sentence)
			if key == "" {
				continue
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			target.Focuses[targetFocus] = append(target.Focuses[targetFocus], entry)
			added++
		}
	}
	return added
}

func collectSentences(pack packFile) []string {
	sentences := []string{}
	for _, entries := range pack.Focuses {
		for _, entry := range entries {
			if strings.TrimSpace(entry.Sentence) != "" {
				sentences = append(sentences, entry.Sentence)
			}
		}
	}
	return sentences
}

func normalizeKey(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeSentence(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
