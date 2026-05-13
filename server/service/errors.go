package service

import "errors"

var errCampaignNotDraft = errors.New("only draft campaigns can be updated")

var errCampaignNotArchivable = errors.New("only draft or published campaigns outside the active period can be archived")

var errCampaignAlreadyArchived = errors.New("campaign is already archived")

var errLandingPageNotDraft = errors.New("only draft landing pages can be updated")

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
