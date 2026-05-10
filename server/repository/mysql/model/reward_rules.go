package model

import "encoding/json"

// RewardRulesPayload is the JSON shape stored in campaigns.reward_rules.
type RewardRulesPayload struct {
	TopupThreshold  float64 `json:"topupThreshold"`
	RewardAmount    float64 `json:"rewardAmount"`
	RewardType      string  `json:"rewardType"`
	MaxClaimPerUser int     `json:"maxClaimPerUser"`
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
