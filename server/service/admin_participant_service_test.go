package service

import (
	"context"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAdminParticipantService_ListAndGet(t *testing.T) {
	joinedAt := time.Unix(12_344, 0).UTC()
	campaigns := &webCampaignRepoStub{campaigns: []model.Campaign{{
		ID: 123, Name: "Summer", BudgetProjectID: 9, Status: model.CampaignStatusPublished,
	}}}
	rules := webRulesRepoStub{rules: []model.CampaignRewardRule{{
		CampaignID: 123, RefClient: model.RewardRefClientTaskGroup, RefID: 77,
	}}}
	parts := &adminParticipantRepoStub{rows: []model.CampaignParticipant{{
		CampaignID: 123, UserID: 1, JoinedAt: joinedAt,
	}}}
	svc := NewAdminParticipantService(campaigns, rules, parts)

	list, err := svc.ListParticipants(123)
	require.NoError(t, err)
	require.Equal(t, int64(123), list.Campaign.ID)
	require.Equal(t, "Summer", list.Campaign.Name)
	require.Equal(t, int64(77), list.Campaign.TaskGroupID)
	require.Equal(t, int64(9), list.Campaign.ProjectID)
	require.Len(t, list.Participants, 1)
	require.Equal(t, int64(1), list.Participants[0].UserID)
	require.Equal(t, int64(12_344), list.Participants[0].JoinedAt)
	require.Equal(t, "LOW", list.Participants[0].RiskLevel)

	detail, err := svc.GetParticipant(123, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), detail.UserID)
	require.Equal(t, int64(12_344), detail.JoinedAt)

	_, err = svc.GetParticipant(123, 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = svc.ListParticipants(999)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAdminParticipantService_ListUnpublishedReturnsEmpty(t *testing.T) {
	campaigns := &webCampaignRepoStub{campaigns: []model.Campaign{{
		ID: 10, Name: "Draft", BudgetProjectID: 9, Status: model.CampaignStatusDraft,
	}}}
	rules := webRulesRepoStub{rules: []model.CampaignRewardRule{{
		CampaignID: 10, RefClient: model.RewardRefClientTaskGroup, RefID: 77,
	}}}
	parts := &adminParticipantRepoStub{rows: []model.CampaignParticipant{{
		CampaignID: 10, UserID: 1, JoinedAt: time.Unix(12_344, 0).UTC(),
	}}}
	svc := NewAdminParticipantService(campaigns, rules, parts)

	list, err := svc.ListParticipants(10)
	require.NoError(t, err)
	require.Equal(t, int64(10), list.Campaign.ID)
	require.Equal(t, "Draft", list.Campaign.Name)
	require.Zero(t, list.Campaign.TaskGroupID)
	require.Zero(t, list.Campaign.ProjectID)
	require.Empty(t, list.Participants)
}

func TestAdminParticipantService_ListWithoutTaskGroupRule(t *testing.T) {
	campaigns := &webCampaignRepoStub{campaigns: []model.Campaign{{
		ID: 5, Name: "NoGroup", BudgetProjectID: 3, Status: model.CampaignStatusPublished,
	}}}
	svc := NewAdminParticipantService(campaigns, webRulesRepoStub{rules: []model.CampaignRewardRule{{
		CampaignID: 5, RefClient: model.RewardRefClientTask, RefID: 9,
	}}}, &adminParticipantRepoStub{})

	list, err := svc.ListParticipants(5)
	require.NoError(t, err)
	require.Zero(t, list.Campaign.TaskGroupID)
	require.Equal(t, int64(3), list.Campaign.ProjectID)
	require.Empty(t, list.Participants)
}

func TestAdminParticipantService_GetRequiresCampaign(t *testing.T) {
	svc := NewAdminParticipantService(
		&webCampaignRepoStub{},
		webRulesRepoStub{},
		&adminParticipantRepoStub{rows: []model.CampaignParticipant{{
			CampaignID: 1, UserID: 1, JoinedAt: time.Unix(1, 0).UTC(),
		}}},
	)
	_, err := svc.GetParticipant(1, 1)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

type adminParticipantRepoStub struct {
	rows []model.CampaignParticipant
}

func (r *adminParticipantRepoStub) GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error) {
	return r.GetByCampaignAndUserContext(context.Background(), campaignID, userID)
}

func (r *adminParticipantRepoStub) GetByCampaignAndUserContext(_ context.Context, campaignID, userID int64) (*model.CampaignParticipant, error) {
	for i := range r.rows {
		if r.rows[i].CampaignID == campaignID && r.rows[i].UserID == userID {
			return &r.rows[i], nil
		}
	}
	return nil, nil
}

func (r *adminParticipantRepoStub) ListByCampaignID(campaignID int64) ([]model.CampaignParticipant, error) {
	out := make([]model.CampaignParticipant, 0)
	for _, row := range r.rows {
		if row.CampaignID == campaignID {
			out = append(out, row)
		}
	}
	return out, nil
}

func (r *adminParticipantRepoStub) ListJoinedCampaignIDs(int64, []int64) (map[int64]struct{}, error) {
	return map[int64]struct{}{}, nil
}

func (r *adminParticipantRepoStub) Join(context.Context, int64, int64) (*model.CampaignParticipant, error) {
	return nil, nil
}
