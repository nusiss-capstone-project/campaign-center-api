package data

// WebCampaignListItem is one campaign in the user list mock response.
type WebCampaignListItem struct {
	ID                int64  `json:"id" example:"1001"`
	Name              string `json:"name" example:"Summer Deposit Bonus"`
	Market            string `json:"market" example:"SG"`
	Status            int16  `json:"status" example:"2"`
	CampaignStartTime int64  `json:"campaignStartTime" example:"1717200000"`
	CampaignEndTime   int64  `json:"campaignEndTime" example:"1722470400"`
	LandingPageID     int64  `json:"landingPageId" example:"2001"`
	Joined            bool   `json:"joined" example:"false"`
}

// WebCampaignListData is the user list campaigns mock payload.
type WebCampaignListData struct {
	Ongoing  []WebCampaignListItem `json:"ongoing"`
	Upcoming []WebCampaignListItem `json:"upcoming"`
}

// WebCampaignLandingPageData is the landing-page UI mock payload.
type WebCampaignLandingPageData struct {
	CampaignID   int64                    `json:"campaignId" example:"1001"`
	UserID       int64                    `json:"userId" example:"42"`
	Name         string                   `json:"name" example:"Summer Deposit Bonus"`
	Market       string                   `json:"market" example:"SG"`
	TimeZone     string                   `json:"timeZone" example:"Asia/Singapore"`
	Joined       bool                     `json:"joined" example:"false"`
	LandingPage  WebLandingPageContent    `json:"landingPage"`
	Participation WebParticipationStatus  `json:"participation"`
}

// WebLandingPageContent is landing copy for the user UI mock.
type WebLandingPageContent struct {
	Lang           string `json:"lang" example:"en"`
	BannerImageURL string `json:"bannerImageUrl" example:"https://cdn.example.com/banner.png"`
	Title          string `json:"title" example:"Deposit and get a bonus"`
	Description    string `json:"description" example:"Join the campaign and complete the deposit task."`
	Terms          string `json:"terms" example:"One reward per user."`
}

// WebParticipationStatus is mock join/task/reward status for the landing UI.
type WebParticipationStatus struct {
	Joined       bool   `json:"joined" example:"false"`
	TaskStatus   string `json:"taskStatus" example:"NOT_STARTED"`
	RewardStatus string `json:"rewardStatus" example:"NOT_GRANTED"`
}

// WebJoinCampaignData is the join mock response.
type WebJoinCampaignData struct {
	CampaignID int64  `json:"campaignId" example:"1001"`
	UserID     int64  `json:"userId" example:"42"`
	Joined     bool   `json:"joined" example:"true"`
	JoinedAt   int64  `json:"joinedAt" example:"1717200000"`
	Message    string `json:"message" example:"joined (mock)"`
}

// WebDepositReq is the deposit mock request body.
type WebDepositReq struct {
	Amount   float64 `json:"amount" binding:"required" example:"100"`
	Currency string  `json:"currency" example:"USDT"`
}

// WebDepositData is the deposit mock response.
type WebDepositData struct {
	CampaignID       int64   `json:"campaignId" example:"1001"`
	UserID           int64   `json:"userId" example:"42"`
	Amount           float64 `json:"amount" example:"100"`
	Currency         string  `json:"currency" example:"USDT"`
	TaskStatus       string  `json:"taskStatus" example:"COMPLETED"`
	RewardStatus     string  `json:"rewardStatus" example:"pending_distribution"`
	LedgerID         int64   `json:"ledgerId" example:"9001"`
	Message          string  `json:"message" example:"deposit accepted (mock)"`
}
