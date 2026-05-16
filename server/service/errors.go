package service

import (
	"errors"

	"github.com/lianjin/campaign-center-api/server/proxy"
)

var errCampaignNotDraft = errors.New("only draft campaigns can be updated")

var errCampaignNotArchivable = errors.New("only draft or published campaigns outside the active period can be archived")

var errCampaignAlreadyArchived = errors.New("campaign is already archived")

var errLandingPageNotDraft = errors.New("only draft landing pages can be updated")

var errTranslationSourceEmpty = errors.New("translation source text is empty")

var errInvalidAccountInput = errors.New("invalid account input")

// IsCampaignNotDraft reports whether err is the draft-only update constraint.
func IsCampaignNotDraft(err error) bool {
	return errors.Is(err, errCampaignNotDraft)
}

// IsCampaignNotArchivable reports archive eligibility errors (wrong status or not yet ended).
func IsCampaignNotArchivable(err error) bool {
	return errors.Is(err, errCampaignNotArchivable)
}

// IsCampaignAlreadyArchived reports when status is already archiveed.
func IsCampaignAlreadyArchived(err error) bool {
	return errors.Is(err, errCampaignAlreadyArchived)
}

// IsLandingPageNotDraft reports whether err is the draft-only update constraint.
func IsLandingPageNotDraft(err error) bool {
	return errors.Is(err, errLandingPageNotDraft)
}

// IsOpenAINotConfigured reports missing OpenAI credentials.
func IsOpenAINotConfigured(err error) bool {
	return errors.Is(err, proxy.ErrOpenAINotConfigured)
}

// IsTranslationSourceEmpty reports empty merged source for LLM.
func IsTranslationSourceEmpty(err error) bool {
	return errors.Is(err, errTranslationSourceEmpty)
}

// IsInvalidAccountInput reports account operation validation failures.
func IsInvalidAccountInput(err error) bool {
	return errors.Is(err, errInvalidAccountInput)
}
