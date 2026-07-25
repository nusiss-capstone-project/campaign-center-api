package service

// API response messages for the {code,message,data} envelope.
// Shared by admin HTTP handlers and user campaign service.

const (
	MsgSuccess = "success"

	// Admin campaign request validation.
	MsgInvalidCampaignID = "invalid campaignId"
	MsgInvalidVersion    = "invalid version"
	MsgCampaignNameRequired = "campaign name is required"

	// Campaign lookup and availability.
	MsgCampaignNotFound         = "campaign not found"
	MsgCampaignDraftNotFound    = "campaign draft version not found"
	MsgCampaignDraftNotEditable = "only draft campaign versions can be edited"
	MsgCampaignNoDraftToPublish = "no campaign draft version to publish"
	MsgCampaignPublishInvalid   = "campaign publish validation failed"
	MsgCampaignNotAvailable     = "campaign not available"

	// Legacy time parse messages (kept for other callers).
	MsgInvalidRegistrationStartTime = "invalid registrationStartTime"
	MsgInvalidRegistrationEndTime   = "invalid registrationEndTime"
	MsgInvalidCampaignStartTime     = "invalid campaignStartTime"
	MsgInvalidCampaignEndTime       = "invalid campaignEndTime"

	// User eligibility and account.
	MsgUserNotEligible       = "User is not eligible for this campaign"
	MsgUserNotFound          = "user not found"
	MsgUserNotJoinedCampaign = "user has not joined this campaign"

	// Landing page.
	MsgLandingPageNotConfigured = "landing page not configured"
	MsgLandingPageNotFound      = "landing page not found"

	// Top-up and rewards.
	MsgRewardAlreadyGranted    = "Reward already granted"
	MsgRewardAlreadyProcessing = "Reward already processing"
	MsgTopupAmountNotQualified = "Top-up amount does not meet campaign requirement"
	MsgManualReviewRequired    = "manual review required"
	MsgRewardProcessing        = "reward processing"
	MsgInvalidRewardModeFmt    = "invalid rewardMode: %s"
	MsgRewardAmountNonNegative = "reward amount must be non-negative"
)
