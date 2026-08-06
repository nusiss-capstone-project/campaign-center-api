package service

import "errors"

// API response messages for the {code,message,data} envelope.

const (
	MsgSuccess = "success"

	// Admin campaign request validation.
	MsgInvalidCampaignID    = "invalid campaignId"
	MsgInvalidVersion       = "invalid version"
	MsgCampaignNameRequired = "campaign name is required"

	// Campaign lookup.
	MsgCampaignNotFound         = "campaign not found"
	MsgCampaignDraftNotFound    = "campaign draft version not found"
	MsgCampaignDraftNotEditable = "only draft campaign versions can be edited"
	MsgCampaignNoDraftToPublish = "no campaign draft version to publish"
	MsgCampaignPublishInvalid   = "campaign publish validation failed"

	MsgUserNotEligible = "user is not eligible for this campaign"
)

// ErrUserNotEligible is returned when MatchUserGroup returns matched=false.
var ErrUserNotEligible = errors.New(MsgUserNotEligible)
