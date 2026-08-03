package service

import (
	"context"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type webCampaignRepoStub struct {
	campaigns []model.Campaign
	getErr    error
}

func (r *webCampaignRepoStub) Create(context.Context, *model.Campaign) error { return nil }
func (r *webCampaignRepoStub) Update(context.Context, *model.Campaign) error { return nil }
func (r *webCampaignRepoStub) Count(mysql.CampaignQuery) (int64, error) {
	return int64(len(r.campaigns)), nil
}
func (r *webCampaignRepoStub) Find(mysql.CampaignQuery, int, int) ([]model.Campaign, error) {
	return r.campaigns, nil
}
func (r *webCampaignRepoStub) UpdateStatus(context.Context, int64, int16, string) (*model.Campaign, error) {
	return nil, nil
}
func (r *webCampaignRepoStub) GetByID(id int64) (*model.Campaign, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for i := range r.campaigns {
		if r.campaigns[i].ID == id {
			return &r.campaigns[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type webPageRepoStub struct {
	page *model.CampaignLandingPage
}

func (r webPageRepoStub) Create(*model.CampaignLandingPage) error  { return nil }
func (r webPageRepoStub) Update(*model.CampaignLandingPage) error  { return nil }
func (r webPageRepoStub) Publish(int64, string) (*model.CampaignLandingPage, error) {
	return nil, nil
}
func (r webPageRepoStub) List(mysql.LandingPageListFilter) ([]model.CampaignLandingPage, int64, error) {
	return nil, 0, nil
}
func (r webPageRepoStub) GetByID(int64) (*model.CampaignLandingPage, error) {
	if r.page == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.page, nil
}

type webTransRepoStub struct {
	row *model.CampaignLandingPageTranslation
}

func (r webTransRepoStub) GetByLandingPageAndLang(int64, string) (*model.CampaignLandingPageTranslation, error) {
	return r.row, nil
}
func (r webTransRepoStub) ListLangsByLandingPageID(int64) ([]string, error) { return nil, nil }
func (r webTransRepoStub) Upsert(*model.CampaignLandingPageTranslation) error { return nil }

type webParticipantRepoStub struct {
	joined map[int64]struct{}
	row    *model.CampaignParticipant
}

func (r *webParticipantRepoStub) GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error) {
	return r.row, nil
}
func (r *webParticipantRepoStub) ListJoinedCampaignIDs(userID int64, campaignIDs []int64) (map[int64]struct{}, error) {
	if r.joined == nil {
		return map[int64]struct{}{}, nil
	}
	return r.joined, nil
}
func (r *webParticipantRepoStub) Join(ctx context.Context, campaignID, userID int64) (*model.CampaignParticipant, error) {
	now := time.Unix(1_700_000_000, 0).UTC()
	return &model.CampaignParticipant{CampaignID: campaignID, UserID: userID, JoinedAt: now}, nil
}

func TestClassifyCampaignWindow(t *testing.T) {
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	startPast := now.Add(-time.Hour)
	startFuture := now.Add(time.Hour)
	endFuture := now.Add(2 * time.Hour)
	endPast := now.Add(-time.Minute)

	require.Equal(t, "upcoming", classifyCampaignWindow(now, &startFuture, &endFuture))
	require.Equal(t, "ongoing", classifyCampaignWindow(now, &startPast, &endFuture))
	require.Equal(t, "ended", classifyCampaignWindow(now, &startPast, &endPast))
	require.Equal(t, "ongoing", classifyCampaignWindow(now, nil, nil))
}

func TestWebCampaignService_ListCampaigns_splitsAndTitles(t *testing.T) {
	now := time.Now().UTC()
	startPast := now.Add(-time.Hour)
	endFuture := now.Add(24 * time.Hour)
	startFuture := now.Add(24 * time.Hour)
	endLater := now.Add(48 * time.Hour)

	campaigns := []model.Campaign{
		{ID: 1, LandingPageID: 10, Status: model.CampaignStatusPublished, Market: "SG",
			CampaignStartTime: &startPast, CampaignEndTime: &endFuture},
		{ID: 2, LandingPageID: 10, Status: model.CampaignStatusPublished, Market: "SG",
			CampaignStartTime: &startFuture, CampaignEndTime: &endLater},
	}
	svc := NewWebCampaignService(
		&webCampaignRepoStub{campaigns: campaigns},
		webPageRepoStub{page: &model.CampaignLandingPage{ID: 10, DefaultLang: "en", Title: "EN Title"}},
		webTransRepoStub{row: &model.CampaignLandingPageTranslation{Title: "ZH Title"}},
		&webParticipantRepoStub{joined: map[int64]struct{}{1: {}}},
	)

	out, err := svc.ListCampaigns(context.Background(), 42, "zh-CN")
	require.NoError(t, err)
	require.Len(t, out.Ongoing, 1)
	require.Len(t, out.Upcoming, 1)
	require.Equal(t, "ZH Title", out.Ongoing[0].Title)
	require.True(t, out.Ongoing[0].Joined)
	require.False(t, out.Upcoming[0].Joined)
}

func TestWebCampaignService_GetCampaignLanding_notPublished(t *testing.T) {
	svc := NewWebCampaignService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, Status: model.CampaignStatusDraft}}},
		webPageRepoStub{},
		webTransRepoStub{},
		&webParticipantRepoStub{},
	)
	_, err := svc.GetCampaignLanding(context.Background(), 1, 42, "en")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestWebCampaignService_GetCampaignLanding_success(t *testing.T) {
	joinedAt := time.Unix(1_700_000_000, 0).UTC()
	svc := NewWebCampaignService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{
			ID: 3, Name: "C", Market: "SG", TimeZone: "Asia/Singapore",
			Status: model.CampaignStatusPublished, LandingPageID: 10,
		}}},
		webPageRepoStub{page: &model.CampaignLandingPage{
			ID: 10, DefaultLang: "en", Title: "EN", Description: "d", Terms: "t", BannerImageURL: "u",
		}},
		webTransRepoStub{row: &model.CampaignLandingPageTranslation{
			Title: "ZH", Description: "zd", Terms: "zt",
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{
			CampaignID: 3, UserID: 42, JoinedAt: joinedAt,
		}},
	)
	out, err := svc.GetCampaignLanding(context.Background(), 3, 42, "zh-CN")
	require.NoError(t, err)
	require.Equal(t, int64(3), out.CampaignID)
	require.True(t, out.Joined)
	require.Equal(t, int64(1_700_000_000), out.JoinedAt)
	require.Equal(t, "ZH", out.LandingPage.Title)
	require.Equal(t, "zh-CN", out.LandingPage.Lang)
}

func TestWebCampaignService_JoinCampaign(t *testing.T) {
	svc := NewWebCampaignService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 9, Status: model.CampaignStatusPublished}}},
		webPageRepoStub{},
		webTransRepoStub{},
		&webParticipantRepoStub{},
	)
	out, err := svc.JoinCampaign(context.Background(), 9, 7)
	require.NoError(t, err)
	require.Equal(t, &data.WebJoinCampaignData{
		CampaignID: 9, UserID: 7, Joined: true, JoinedAt: 1_700_000_000, Message: "joined",
	}, out)
}
