package service

// API response messages for the {code,message,data} envelope.
// Shared by admin HTTP handlers and user campaign service.

const (
	MsgSuccess = "success"

	// Admin campaign request validation.
	MsgInvalidCampaignID              = "invalid campaignId"
	MsgInvalidRegistrationStartTime = "invalid registrationStartTime"
	MsgInvalidRegistrationEndTime   = "invalid registrationEndTime"
	MsgInvalidCampaignStartTime     = "invalid campaignStartTime"
	MsgInvalidCampaignEndTime       = "invalid campaignEndTime"

	// Campaign lookup and availability.
	MsgCampaignNotFound     = "campaign not found"
	MsgCampaignNotAvailable = "campaign not available"

	// User eligibility and account.
	MsgUserNotEligible       = "User is not eligible for this campaign"
	MsgUserNotFound          = "user not found"
	MsgUserNotJoinedCampaign = "user has not joined this campaign"

	// Landing page.
	MsgLandingPageNotConfigured = "landing page not configured"
	MsgLandingPageNotFound      = "landing page not found"

	// Top-up and rewards.
	MsgRewardAlreadyGranted      = "Reward already granted"
	MsgRewardAlreadyProcessing   = "Reward already processing"
	MsgTopupAmountNotQualified   = "Top-up amount does not meet campaign requirement"
	MsgManualReviewRequired      = "manual review required"
	MsgRewardProcessing          = "reward processing"
	MsgInvalidRewardModeFmt      = "invalid rewardMode: %s"
	MsgRewardAmountNonNegative   = "reward amount must be non-negative"
)
