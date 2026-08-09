package data

// AdminParticipantCampaignVO is the campaign summary on participant list.
type AdminParticipantCampaignVO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TaskGroupID int64  `json:"task_group_id"`
	ProjectID   int64  `json:"project_id"`
}

// AdminParticipantVO is one participant row for admin APIs.
type AdminParticipantVO struct {
	UserID    int64  `json:"user_id"`
	JoinedAt  int64  `json:"joined_at"`
	RiskLevel string `json:"risk_level"`
}

// AdminParticipantListData is StandardResponse.data for participant list.
type AdminParticipantListData struct {
	Campaign     AdminParticipantCampaignVO `json:"campaign"`
	Participants []AdminParticipantVO       `json:"participants"`
}
