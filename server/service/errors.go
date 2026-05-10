package service

import "errors"

var errCampaignNotDraft = errors.New("only draft campaigns can be updated")

var errLandingPageNotDraft = errors.New("only draft landing pages can be updated")

// IsCampaignNotDraft reports whether err is the draft-only update constraint.
func IsCampaignNotDraft(err error) bool {
	return errors.Is(err, errCampaignNotDraft)
}

// IsLandingPageNotDraft reports whether err is the draft-only update constraint.
func IsLandingPageNotDraft(err error) bool {
	return errors.Is(err, errLandingPageNotDraft)
}
