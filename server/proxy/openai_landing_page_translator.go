package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/lianjin/campaign-center-api/server/config"
	"github.com/lianjin/campaign-center-api/server/log"
)

const landingTranslateSystemPrompt = `You are a professional multilingual translation assistant for marketing campaign content.
Translate the provided content from the source language to the target language.
Requirements:
1. Preserve all placeholders exactly as-is (e.g. {{amount}}, {{reward_amount}}, {{days}}).
2. Do NOT translate placeholders.
3. Do NOT modify placeholder format.
4. Keep the response concise and natural for marketing usage.
5. Return valid JSON only with keys title, description, terms.
6. Do not add explanations.`

type openAILandingPageTranslator struct {
	client *openai.Client
	model  string
}

var (
	landingPageTranslatorOnce sync.Once
	landingPageTranslatorInst LandingPageTranslator
)

// GetLandingPageTranslator returns a singleton translator (OpenAI or not-configured stub).
func GetLandingPageTranslator() LandingPageTranslator {
	landingPageTranslatorOnce.Do(func() {
		landingPageTranslatorInst = newTranslatorFromConfig()
	})
	return landingPageTranslatorInst
}

func newTranslatorFromConfig() LandingPageTranslator {
	cfg := config.Config
	if cfg == nil || cfg.OpenAIConfig == nil || strings.TrimSpace(cfg.OpenAIConfig.APIKey) == "" {
		return notConfiguredTranslator{}
	}
	return newOpenAILandingPageTranslator(cfg.OpenAIConfig)
}

type notConfiguredTranslator struct{}

func (notConfiguredTranslator) Translate(context.Context, LandingPageTranslateInput) (*LandingPageTranslateOutput, error) {
	return nil, ErrOpenAINotConfigured
}

func newOpenAILandingPageTranslator(oc *config.OpenAIConfig) *openAILandingPageTranslator {
	cc := openai.DefaultConfig(strings.TrimSpace(oc.APIKey))
	if strings.TrimSpace(oc.BaseURL) != "" {
		cc.BaseURL = strings.TrimRight(strings.TrimSpace(oc.BaseURL), "/") + "/v1"
	}
	model := strings.TrimSpace(oc.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &openAILandingPageTranslator{client: openai.NewClientWithConfig(cc), model: model}
}

func (t *openAILandingPageTranslator) Translate(ctx context.Context, in LandingPageTranslateInput) (*LandingPageTranslateOutput, error) {
	log.Logger.Infow("openai_translate_request", "target_lang", in.TargetLang, "source_lang", in.SourceLang)
	out, err := t.callChat(ctx, in)
	if err != nil {
		log.Logger.Errorw("openai_translate_failed", "error", err)
		return nil, err
	}
	log.Logger.Infow("openai_translate_success", "target_lang", in.TargetLang)
	return out, nil
}

func (t *openAILandingPageTranslator) callChat(ctx context.Context, in LandingPageTranslateInput) (*LandingPageTranslateOutput, error) {
	userJSON, err := marshalUserPrompt(in)
	if err != nil {
		return nil, err
	}
	req := openai.ChatCompletionRequest{
		Model: t.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: landingTranslateSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userJSON},
		},
		Temperature: 0.2,
	}
	resp, err := t.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("openai: empty choices")
	}
	return parseTranslateJSON(resp.Choices[0].Message.Content)
}

func marshalUserPrompt(in LandingPageTranslateInput) (string, error) {
	payload := map[string]any{
		"source_lang": in.SourceLang,
		"target_lang": in.TargetLang,
		"content": map[string]string{
			"title":       in.Title,
			"description": in.Description,
			"terms":       in.Terms,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseTranslateJSON(raw string) (*LandingPageTranslateOutput, error) {
	s := stripJSONFence(raw)
	var parsed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Terms       string `json:"terms"`
	}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return nil, err
	}
	return &LandingPageTranslateOutput{
		Title: parsed.Title, Description: parsed.Description, Terms: parsed.Terms,
	}, nil
}

func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[idx+1:]
	}
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
