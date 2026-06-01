package data

import "errors"

var (
	// ErrCampaignNotDraft is returned when a non-draft campaign is updated.
	ErrCampaignNotDraft = errors.New("only draft campaigns can be updated")
	// ErrCampaignNotArchivable is returned when archive preconditions fail.
	ErrCampaignNotArchivable = errors.New("only draft or published campaigns outside the active period can be archived")
	// ErrCampaignAlreadyArchived is returned when status is already archived.
	ErrCampaignAlreadyArchived = errors.New("campaign is already archived")
	// ErrLandingPageNotDraft is returned when a non-draft landing page is updated.
	ErrLandingPageNotDraft = errors.New("only draft landing pages can be updated")
	// ErrTranslationSourceEmpty is returned when merged LLM source text is empty.
	ErrTranslationSourceEmpty = errors.New("translation source text is empty")
	// ErrInvalidAccountInput is returned for account operation validation failures.
	ErrInvalidAccountInput = errors.New("invalid account input")
)

// IsCampaignNotDraft reports whether err is the draft-only update constraint.
func IsCampaignNotDraft(err error) bool {
	return errors.Is(err, ErrCampaignNotDraft)
}

// IsCampaignNotArchivable reports archive eligibility errors (wrong status or not yet ended).
func IsCampaignNotArchivable(err error) bool {
	return errors.Is(err, ErrCampaignNotArchivable)
}

// IsCampaignAlreadyArchived reports when status is already archived.
func IsCampaignAlreadyArchived(err error) bool {
	return errors.Is(err, ErrCampaignAlreadyArchived)
}

// IsLandingPageNotDraft reports whether err is the draft-only update constraint.
func IsLandingPageNotDraft(err error) bool {
	return errors.Is(err, ErrLandingPageNotDraft)
}

// IsTranslationSourceEmpty reports empty merged source for LLM.
func IsTranslationSourceEmpty(err error) bool {
	return errors.Is(err, ErrTranslationSourceEmpty)
}

// IsInvalidAccountInput reports account operation validation failures.
func IsInvalidAccountInput(err error) bool {
	return errors.Is(err, ErrInvalidAccountInput)
}
