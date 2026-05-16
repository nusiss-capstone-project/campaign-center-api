package model

import "encoding/json"

// RewardRulesPayload is the JSON shape stored in campaigns.reward_rules.
type RewardRulesPayload struct {
	TopupThreshold float64 `json:"topupThreshold"`
	RewardType     string  `json:"rewardType"`

	// compatible with existing data
	RewardAmount   float64 `json:"rewardAmount,omitempty"`
	RewardCurrency string  `json:"rewardCurrency,omitempty"`

	// new fields
	RewardMode       string  `json:"rewardMode"` // FIXED_AMOUNT / PERCENTAGE
	RewardPercentage float64 `json:"rewardPercentage,omitempty"`
	MaxRewardAmount  float64 `json:"maxRewardAmount,omitempty"`

	MaxClaimPerUser int `json:"maxClaimPerUser"`
	MinObtainDays   int `json:"minObtainDays"` // todo implement, now reward immediately
}

func ParseRewardRulesJSON(s string) (RewardRulesPayload, error) {
	var out RewardRulesPayload
	if s == "" {
		return out, nil
	}
	err := json.Unmarshal([]byte(s), &out)
	return out, err
}

func MarshalRewardRulesPayload(r RewardRulesPayload) (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
