package data

import "errors"

var (
	// ErrCampaignDraftNotEditable is returned when editing a non-draft version.
	ErrCampaignDraftNotEditable = errors.New("only draft campaign versions can be edited")
	// ErrCampaignNoDraftToPublish is returned when publish finds no draft version.
	ErrCampaignNoDraftToPublish = errors.New("no campaign draft version to publish")
	// ErrCampaignPublishInvalid is returned when publish payload fails required checks.
	ErrCampaignPublishInvalid = errors.New("campaign publish validation failed")
	// ErrLandingPageNotDraft is returned when a non-draft landing page is updated.
	ErrLandingPageNotDraft = errors.New("only draft landing pages can be updated")
	// ErrTranslationSourceEmpty is returned when merged LLM source text is empty.
	ErrTranslationSourceEmpty = errors.New("translation source text is empty")
	// ErrInvalidAccountInput is returned for account operation validation failures.
	ErrInvalidAccountInput = errors.New("invalid account input")
)

// IsCampaignDraftNotEditable reports draft-only version edit constraint.
func IsCampaignDraftNotEditable(err error) bool {
	return errors.Is(err, ErrCampaignDraftNotEditable)
}

// IsCampaignNoDraftToPublish reports missing draft on publish.
func IsCampaignNoDraftToPublish(err error) bool {
	return errors.Is(err, ErrCampaignNoDraftToPublish)
}

// IsCampaignPublishInvalid reports publish validation failures.
func IsCampaignPublishInvalid(err error) bool {
	return errors.Is(err, ErrCampaignPublishInvalid)
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
