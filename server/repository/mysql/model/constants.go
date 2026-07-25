package model

// Campaign status (smallint): 1 draft, 2 published
const (
	CampaignStatusDraft     int16 = 1
	CampaignStatusPublished int16 = 2
)

// Campaign draft status (varchar on campaign_drafts).
const (
	CampaignDraftStatusDraft     = "draft"
	CampaignDraftStatusPublished = "published"
)

// Campaign reward rule ref_client values.
const (
	RewardRefClientTask      = "task"
	RewardRefClientTaskGroup = "task_group"
)

// Campaign user reward ledger statuses.
const (
	LedgerRewardStatusPendingDistribution = "pending_distribution"
	LedgerRewardStatusDistributing        = "distributing"
	LedgerRewardStatusDistributeSuccess   = "distribute_success"
	LedgerRewardStatusDistributeFail      = "distribute_fail"
)

// Landing page status
const (
	LandingPageStatusDraft     int16 = 1
	LandingPageStatusPublished int16 = 2
	LandingPageStatusArchive   int16 = 3
)

const (
	MarketGlobal = "GLOBAL"
	MarketUS     = "US"
	MarketEU     = "EU"
	MarketSEA    = "SEA"
	MarketHK     = "HK"
	MarketJP     = "JP"
	MarketSG     = "SG"
	MarketEEA    = "EEA"
	MarketTR     = "TR"
	MarketBR     = "BR"
)

const (
	KYCStatusPassed  = "PASSED"
	KYCStatusPending = "PENDING"
	KYCStatusFailed  = "FAILED"
)

const (
	JoinStatusJoined = "JOINED"
)

const (
	TaskStatusNotStarted   = "NOT_STARTED"
	TaskStatusCompleted    = "COMPLETED"
	TaskStatusNotQualified = "NOT_QUALIFIED"
)

const (
	RewardStatusNotGranted    = "NOT_GRANTED"
	RewardStatusPending       = "PENDING"
	RewardStatusGranted       = "GRANTED"
	RewardStatusPendingReview = "PENDING_REVIEW"
)

const (
	RiskStatusApproved     = "APPROVED"
	RiskStatusManualReview = "MANUAL_REVIEW"
	RiskStatusRejected     = "REJECTED"
)

const (
	RewardTypeBonusCredit      = "BONUS_CREDIT"
	RewardTypeTokenBonus       = "TOKEN_BONUS"
	RewardTypeTradingFeeRebate = "TRADING_FEE_REBATE"
	RewardTypeVoucher          = "VOUCHER"
	RewardTypeMysteryBox       = "MYSTERY_BOX"
	RewardTypePoints           = "POINTS"
)

const (
	RewardModeFixedAmount = "FIXED_AMOUNT"
	RewardModePercentage  = "PERCENTAGE"
)

const (
	RewardTxnStatusCompleted = "COMPLETED"
	RewardTxnStatusPending   = "PENDING"
)

const (
	RiskLevelHigh = "HIGH"
)

const (
	RejectReasonKYCNotPassed = "KYC_NOT_PASSED"
	RejectReasonSegment      = "SEGMENT_MISMATCH"
)

const DefaultCurrency = "USDT"

const (
	AccountTxnTypeRecharge        = "RECHARGE"
	AccountTxnTypeCampaignReward  = "CAMPAIGN_REWARD"
	AccountTxnStatusSuccess       = "SUCCESS"
	AccountTxnStatusFailed        = "FAILED"
	AccountTxnRelatedTypeCampaign = "CAMPAIGN"
	AccountTxnRelatedTypeRecharge = "RECHARGE"
)
