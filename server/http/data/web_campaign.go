package data

// WebCampaignListItem is one campaign on the user list page.
type WebCampaignListItem struct {
	ID                int64  `json:"id" example:"1001"`
	Title             string `json:"title" example:"Deposit and get a bonus"`
	Market            string `json:"market" example:"SG"`
	Status            int16  `json:"status" example:"2"`
	CampaignStartTime int64  `json:"campaignStartTime" example:"1717200000"`
	CampaignEndTime   int64  `json:"campaignEndTime" example:"1722470400"`
	LandingPageID     int64  `json:"landingPageId" example:"2001"`
	Joined            bool   `json:"joined" example:"false"`
}

// WebCampaignListData is the user list campaigns payload (published only).
type WebCampaignListData struct {
	Ongoing  []WebCampaignListItem `json:"ongoing"`
	Upcoming []WebCampaignListItem `json:"upcoming"`
}

// WebCampaignLandingPageData is the campaign detail / landing-page payload.
type WebCampaignLandingPageData struct {
	CampaignID    int64                 `json:"campaignId" example:"1001"`
	UserID        int64                 `json:"userId" example:"42"`
	Name          string                `json:"name" example:"Summer Deposit Bonus"`
	Market        string                `json:"market" example:"SG"`
	TimeZone      string                `json:"timeZone" example:"Asia/Singapore"`
	Joined        bool                  `json:"joined" example:"false"`
	JoinedAt      int64                 `json:"joinedAt,omitempty" example:"1717200000"`
	LandingPage   WebLandingPageContent `json:"landingPage"`
}

// WebLandingPageContent is localized landing copy for the user UI.
type WebLandingPageContent struct {
	Lang           string                        `json:"lang" example:"en"`
	BannerImageURL string                        `json:"bannerImageUrl" example:"https://cdn.example.com/banner.png"`
	Title          string                        `json:"title" example:"Deposit and get a bonus"`
	Description    string                        `json:"description" example:"Join the campaign and complete the deposit task."`
	Terms          string                        `json:"terms" example:"One reward per user."`
	Steps          []LandingPageRepeatableItemVO `json:"steps"`
	Faq            []LandingPageRepeatableItemVO `json:"faq"`
}

// WebJoinCampaignData is the join response.
type WebJoinCampaignData struct {
	CampaignID int64  `json:"campaignId" example:"1001"`
	UserID     int64  `json:"userId" example:"42"`
	Joined     bool   `json:"joined" example:"true"`
	JoinedAt   int64  `json:"joinedAt" example:"1717200000"`
	Message    string `json:"message" example:"joined"`
}
