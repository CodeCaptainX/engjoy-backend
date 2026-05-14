package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sentenceminer/internal/sentences/model"
)

type Client struct {
	apiKey   string
	model    string
	ttsModel string
	baseURL  string
	http     *http.Client
}

func NewClient(apiKey, model, ttsModel, baseURL string) *Client {
	return &Client{
		apiKey:   apiKey,
		model:    model,
		ttsModel: ttsModel,
		baseURL:  baseURL,
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) ListModels(ctx context.Context) ([]model.ModelInfo, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is not configured")
	}

	url := fmt.Sprintf("%s/models?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini error: status %d - %s", resp.StatusCode, string(body))
	}

	var out struct {
		Models []model.ModelInfo `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Models, nil
}

func (c *Client) AnalyzeSentence(ctx context.Context, sentence string) (string, error) {
	prompt := "Explain the sentence, give one concise grammar focus, list vocabulary with meanings, and give one example. Respond as pure JSON only with no markdown, no code blocks. Keys: explanation, grammar_focus, vocabulary (array of {word, meaning}), example."
	reqBody := model.GenerateRequest{
		Contents: []model.Content{{
			Role:  "user",
			Parts: []model.Part{{Text: prompt + "\nSentence: " + sentence}},
		}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini error: status %d - %s", resp.StatusCode, string(body))
	}
	var out model.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini empty response")
	}

	text := out.Candidates[0].Content.Parts[0].Text

	// Strip markdown code blocks if Gemini wraps response
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	return text, nil
}

func (c *Client) GenerateCategorySentences(ctx context.Context, category, focus string, existing []string, count int) ([]string, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is not configured")
	}

	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20
	}

	avoidList := ""
	if len(existing) > 0 {
		trimmed := existing
		if len(trimmed) > 40 {
			trimmed = trimmed[:40]
		}
		avoidList = "\nAvoid generating these existing sentences or close paraphrases:\n- " + strings.Join(trimmed, "\n- ")
	}

	focusInstruction := "Use a balanced mix of common subtopics inside this category."
	cleanFocus := strings.TrimSpace(strings.ToLower(focus))
	if cleanFocus != "" && cleanFocus != "all" {
		focusInstruction = fmt.Sprintf("Focus specifically on the subtopic \"%s\" inside this category.", focus)
	}

	prompt := fmt.Sprintf(
		"Generate %d natural English learner-friendly sentences for the category \"%s\". "+
			"%s "+
			"Return pure JSON only with this shape: {\"sentences\":[\"...\"]}. "+
			"Each sentence must be unique, practical, and under 110 characters. "+
			"Do not number them. Do not include explanations.%s",
		count,
		category,
		focusInstruction,
		avoidList,
	)

	reqBody := model.GenerateRequest{
		Contents: []model.Content{{
			Role:  "user",
			Parts: []model.Part{{Text: prompt}},
		}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini error: status %d - %s", resp.StatusCode, string(body))
	}

	var out model.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini empty response")
	}

	text := strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var parsed struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(parsed.Sentences))
	seen := make(map[string]struct{}, len(parsed.Sentences))
	for _, sentence := range parsed.Sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, trimmed)
	}

	return results, nil
}

func (c *Client) GenerateSentencePack(ctx context.Context, category, focus string, existing []string, count int) (string, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return "", fmt.Errorf("gemini api key is not configured")
	}

	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}

	avoidList := ""
	if len(existing) > 0 {
		trimmed := existing
		if len(trimmed) > 80 {
			trimmed = trimmed[:80]
		}
		avoidList = "\nAvoid these existing sentences or close paraphrases:\n- " + strings.Join(trimmed, "\n- ")
	}

	prompt := fmt.Sprintf(
		"Generate %d original natural English learner sentences for category %q and focus %q. "+
			"Return pure JSON only, no markdown, with this exact shape: "+
			"{\"category\":\"%s\",\"focuses\":{\"%s\":[{\"sentence\":\"...\",\"meaning\":\"...\",\"grammar_focus\":\"...\",\"vocabulary\":[{\"word\":\"...\",\"meaning\":\"...\"}],\"example\":\"...\"}]}}. "+
			"Each sentence must be practical, under 110 characters, and useful for real conversation. "+
			"Meaning must be learner-friendly. Grammar focus must be concise. Vocabulary should include 2 useful items. Example should show similar usage.%s",
		count,
		category,
		focus,
		category,
		focus,
		avoidList,
	)

	reqBody := model.GenerateRequest{
		Contents: []model.Content{{
			Role:  "user",
			Parts: []model.Part{{Text: prompt}},
		}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini error: status %d - %s", resp.StatusCode, string(body))
	}

	var out model.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini empty response")
	}

	text := strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text), nil
}

func (c *Client) GenerateSpeech(ctx context.Context, sentence, mode string) ([]byte, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is not configured")
	}

	prompt := buildTTSPrompt(sentence, mode)
	reqBody := model.GenerateRequest{
		Contents: []model.Content{{
			Parts: []model.Part{{Text: prompt}},
		}},
		GenerationConfig: &model.GenerationConfig{
			ResponseModalities: []string{"AUDIO"},
			SpeechConfig: &model.SpeechConfig{
				VoiceConfig: &model.VoiceConfig{
					PrebuiltVoiceConfig: &model.PrebuiltVoiceConfig{
						VoiceName: "Sulafat",
					},
				},
			},
		},
		Model: c.ttsModel,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, c.ttsModel)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini tts error: status %d - %s", resp.StatusCode, string(body))
	}

	var out model.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini tts empty response")
	}

	audioData := out.Candidates[0].Content.Parts[0].InlineData
	if audioData == nil || audioData.Data == "" {
		return nil, fmt.Errorf("gemini tts returned no audio data")
	}

	pcm, err := base64.StdEncoding.DecodeString(audioData.Data)
	if err != nil {
		return nil, err
	}

	return pcmToWav(pcm, 24000, 1, 16), nil
}

func buildTTSPrompt(sentence, mode string) string {
	trimmed := strings.TrimSpace(sentence)
	if mode == "slow" {
		return fmt.Sprintf(
			`Read this sentence for an English learner in a warm, clean, enjoyable voice. Speak a little slower than normal, articulate each word clearly, and keep the sentence natural instead of robotic.

Sentence: "%s"`,
			trimmed,
		)
	}

	return fmt.Sprintf(
		`Read this sentence for an English learner in a warm, clean, enjoyable voice. Keep the pronunciation crisp, natural, and easy to follow.

Sentence: "%s"`,
		trimmed,
	)
}

func pcmToWav(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8
	dataLen := len(pcm)
	buffer := bytes.NewBuffer(make([]byte, 0, 44+dataLen))

	buffer.WriteString("RIFF")
	_ = binary.Write(buffer, binary.LittleEndian, uint32(36+dataLen))
	buffer.WriteString("WAVE")
	buffer.WriteString("fmt ")
	_ = binary.Write(buffer, binary.LittleEndian, uint32(16))
	_ = binary.Write(buffer, binary.LittleEndian, uint16(1))
	_ = binary.Write(buffer, binary.LittleEndian, uint16(channels))
	_ = binary.Write(buffer, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buffer, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(buffer, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(buffer, binary.LittleEndian, uint16(bitsPerSample))
	buffer.WriteString("data")
	_ = binary.Write(buffer, binary.LittleEndian, uint32(dataLen))
	buffer.Write(pcm)

	return buffer.Bytes()
}
