package proxy

import (
	"context"
	"errors"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
)

// ErrOpenAINotConfigured is returned when OpenAI credentials are missing.
var ErrOpenAINotConfigured = errors.New("openai is not configured")

// LandingPageTranslator calls an LLM to translate landing page copy.
type LandingPageTranslator interface {
	Translate(ctx context.Context, in LandingPageTranslateInput) (*LandingPageTranslateOutput, error)
}

// LandingPageTranslateInput maps to the design doc user prompt payload.
type LandingPageTranslateInput struct {
	SourceLang  string
	TargetLang  string
	Title       string
	Description string
	Terms       string
	Steps       []data.LandingPageRepeatableItemVO
	Faq         []data.LandingPageRepeatableItemVO
}

// LandingPageTranslateOutput is parsed JSON from the model.
type LandingPageTranslateOutput struct {
	Title       string
	Description string
	Terms       string
	Steps       []data.LandingPageRepeatableItemVO
	Faq         []data.LandingPageRepeatableItemVO
}
