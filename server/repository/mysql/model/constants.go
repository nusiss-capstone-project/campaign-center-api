package model

// Campaign status (smallint): 1 draft, 2 published, 3 archive
const (
	CampaignStatusDraft     int16 = 1
	CampaignStatusPublished int16 = 2
	CampaignStatusArchive   int16 = 3
)

// Landing page status
const (
	LandingPageStatusDraft     int16 = 1
	LandingPageStatusPublished int16 = 2
	LandingPageStatusArchive   int16 = 3
)

const (
	CampaignTypeTopupReward = "TOPUP_REWARD"
)

const (
	MarketGlobal = "GLOBAL"
	MarketUS     = "US"
	MarketEU     = "EU"
	MarketSEA    = "SEA"
	MarketHK     = "HK"
	MarketJP     = "JP"
	MarketSG     = "SG"
)

const (
	UserSegmentNewUser            = "NEW_USER"
	UserSegmentVIPUser            = "VIP_USER"
	UserSegmentInactiveUser       = "INACTIVE_USER"
	UserSegmentHighNetWorth       = "HIGH_NET_WORTH"
	UserSegmentFirstTimeDepositor = "FIRST_TIME_DEPOSITOR"
	UserSegmentAllUsers           = "ALL_USERS"
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
