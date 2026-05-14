package model

type GenerateRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	Model            string            `json:"model,omitempty"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type GenerateResponse struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
}

type ModelInfo struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"displayName"`
	SupportedActions []string `json:"supportedActions"`
}

type GenerationConfig struct {
	ResponseModalities []string      `json:"responseModalities,omitempty"`
	SpeechConfig       *SpeechConfig `json:"speechConfig,omitempty"`
}

type SpeechConfig struct {
	VoiceConfig *VoiceConfig `json:"voiceConfig,omitempty"`
}

type VoiceConfig struct {
	PrebuiltVoiceConfig *PrebuiltVoiceConfig `json:"prebuiltVoiceConfig,omitempty"`
}

type PrebuiltVoiceConfig struct {
	VoiceName string `json:"voiceName"`
}

type InlineData struct {
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}
