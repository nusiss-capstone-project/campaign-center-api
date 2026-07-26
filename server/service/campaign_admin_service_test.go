package service_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type memCampaignRepo struct {
	byID   map[int64]*model.Campaign
	nextID int64
}

func newMemCampaignRepo() *memCampaignRepo {
	return &memCampaignRepo{byID: map[int64]*model.Campaign{}, nextID: 1}
}

func (r *memCampaignRepo) Create(_ context.Context, c *model.Campaign) error {
	c.ID = r.nextID
	r.nextID++
	cp := *c
	r.byID[c.ID] = &cp
	return nil
}

func (r *memCampaignRepo) Update(_ context.Context, c *model.Campaign) error {
	if _, ok := r.byID[c.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *c
	r.byID[c.ID] = &cp
	return nil
}

func (r *memCampaignRepo) GetByID(id int64) (*model.Campaign, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *memCampaignRepo) matching(q mysql.CampaignQuery) []model.Campaign {
	out := make([]model.Campaign, 0, len(r.byID))
	for _, c := range r.byID {
		if q.CampaignID != nil && c.ID != *q.CampaignID {
			continue
		}
		if q.Status != nil && c.Status != *q.Status {
			continue
		}
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (r *memCampaignRepo) Count(q mysql.CampaignQuery) (int64, error) {
	return int64(len(r.matching(q))), nil
}

func (r *memCampaignRepo) Find(q mysql.CampaignQuery, offset, limit int) ([]model.Campaign, error) {
	rows := r.matching(q)
	if offset >= len(rows) {
		return nil, nil
	}
	rows = rows[offset:]
	if limit > 0 && limit < len(rows) {
		rows = rows[:limit]
	}
	return rows, nil
}

func (r *memCampaignRepo) UpdateStatus(ctx context.Context, id int64, status int16, operator string) (*model.Campaign, error) {
	c, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	c.Status = status
	c.UpdatedBy = operator
	c.UpdatedAt = time.Now()
	_ = r.Update(ctx, c)
	return c, nil
}

type memDraftRepo struct {
	rows   []model.CampaignDraft
	nextID int64
}

func newMemDraftRepo() *memDraftRepo {
	return &memDraftRepo{nextID: 1}
}

func (r *memDraftRepo) Create(_ context.Context, d *model.CampaignDraft) error {
	d.ID = r.nextID
	r.nextID++
	r.rows = append(r.rows, *d)
	return nil
}

func (r *memDraftRepo) Update(_ context.Context, d *model.CampaignDraft) error {
	for i := range r.rows {
		if r.rows[i].ID == d.ID {
			r.rows[i] = *d
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (r *memDraftRepo) GetByActivityAndVersion(activityID int64, version int) (*model.CampaignDraft, error) {
	for i := range r.rows {
		if r.rows[i].ActivityID == activityID && r.rows[i].Version == version {
			cp := r.rows[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *memDraftRepo) GetLatestByActivityID(activityID int64) (*model.CampaignDraft, error) {
	var latest *model.CampaignDraft
	for i := range r.rows {
		if r.rows[i].ActivityID != activityID {
			continue
		}
		if latest == nil || r.rows[i].Version > latest.Version {
			cp := r.rows[i]
			latest = &cp
		}
	}
	if latest == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return latest, nil
}

func (r *memDraftRepo) MaxVersion(activityID int64) (int, error) {
	max := 0
	for i := range r.rows {
		if r.rows[i].ActivityID == activityID && r.rows[i].Version > max {
			max = r.rows[i].Version
		}
	}
	return max, nil
}

func (r *memDraftRepo) MaxVersionsByActivityIDs(activityIDs []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(activityIDs))
	for _, id := range activityIDs {
		max, err := r.MaxVersion(id)
		if err != nil {
			return nil, err
		}
		if max > 0 {
			out[id] = max
		}
	}
	return out, nil
}

type memRuleRepo struct {
	byCampaign map[int64][]model.CampaignRewardRule
}

func newMemRuleRepo() *memRuleRepo {
	return &memRuleRepo{byCampaign: map[int64][]model.CampaignRewardRule{}}
}

func (r *memRuleRepo) ListByCampaignID(campaignID int64) ([]model.CampaignRewardRule, error) {
	return append([]model.CampaignRewardRule(nil), r.byCampaign[campaignID]...), nil
}

func (r *memRuleRepo) DeleteByCampaignID(_ context.Context, campaignID int64) error {
	delete(r.byCampaign, campaignID)
	return nil
}

func (r *memRuleRepo) ReplaceByCampaignID(_ context.Context, campaignID int64, rules []model.CampaignRewardRule) error {
	cp := make([]model.CampaignRewardRule, len(rules))
	copy(cp, rules)
	r.byCampaign[campaignID] = cp
	return nil
}

func newAdminSvc() (service.CampaignAdminService, *memCampaignRepo, *memDraftRepo, *memRuleRepo) {
	campaigns := newMemCampaignRepo()
	drafts := newMemDraftRepo()
	rules := newMemRuleRepo()
	return service.NewCampaignAdminService(campaigns, drafts, rules), campaigns, drafts, rules
}

func TestCampaignAdmin_CreateCampaign_nameOnly(t *testing.T) {
	svc, campaigns, drafts, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "  Summer Up  ")
	require.NoError(t, err)
	require.Equal(t, int64(1), id)
	require.Equal(t, model.CampaignStatusDraft, campaigns.byID[id].Status)
	require.Len(t, drafts.rows, 1)
	require.Equal(t, 1, drafts.rows[0].Version)
	require.Equal(t, id, drafts.rows[0].ActivityID)
	require.Equal(t, model.CampaignDraftStatusDraft, drafts.rows[0].Status)
	require.Contains(t, drafts.rows[0].Content, "Summer Up")
}

func TestCampaignAdmin_CreateCampaign_emptyName(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	_, err := svc.CreateCampaign(context.Background(), "   ")
	require.Error(t, err)
}

func TestCampaignAdmin_CreateVersion_requiresCampaign(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	_, err := svc.CreateVersion(context.Background(), 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCampaignAdmin_CreateEditPublishFlow(t *testing.T) {
	svc, _, drafts, rules := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "Summer")
	require.NoError(t, err)

	content := data.CampaignVO{
		Name: "Summer", Market: "SG",
		RegistrationStartTime: 1, RegistrationEndTime: 2,
		CampaignStartTime: 3, CampaignEndTime: 4,
		TimeZone:         "Asia/Singapore",
		TargetUserGroups: data.TargetUserGroupVO{ID: 10, GroupName: "VIP"},
		Budget:           data.BudgetVO{ProjectID: 20, ProjectName: "P1"},
		RewardRules: data.CampaignRewardRuleVO{
			TaskGroupID: 100, TaskGroupReward: 200,
			TaskRewardItems: []data.TaskRewardItemVO{
				{TaskID: 1, TaskName: "t1", RewardTemplateID: 11, RewardTemplateName: "r1"},
			},
		},
		LandingPageID: 7,
	}
	require.NoError(t, svc.EditVersion(context.Background(), id, 1, content))

	detail, err := svc.PublishCampaign(context.Background(), id, "admin")
	require.NoError(t, err)
	require.Equal(t, model.CampaignStatusPublished, detail.Status)
	require.Equal(t, "SG", detail.Market)
	require.Equal(t, int64(20), detail.Budget.ProjectID)
	require.Equal(t, "P1", detail.Budget.ProjectName)
	require.Equal(t, "VIP", detail.TargetUserGroups.GroupName)
	require.Equal(t, int64(100), detail.RewardRules.TaskGroupID)
	require.Equal(t, int64(200), detail.RewardRules.TaskGroupReward)
	require.Len(t, detail.RewardRules.TaskRewardItems, 1)
	require.Equal(t, "t1", detail.RewardRules.TaskRewardItems[0].TaskName)
	require.Equal(t, "r1", detail.RewardRules.TaskRewardItems[0].RewardTemplateName)
	require.Equal(t, 1, drafts.rows[0].Version)
	require.Equal(t, model.CampaignDraftStatusPublished, drafts.rows[0].Status)
	require.Len(t, rules.byCampaign[id], 2)
}

func TestCampaignAdmin_EditPublishedVersion_conflict(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "X")
	require.NoError(t, err)
	content := data.CampaignVO{
		Name: "X", Budget: data.BudgetVO{ProjectID: 1, ProjectName: "p"},
		RewardRules: data.CampaignRewardRuleVO{
			TaskGroupID: 1, TaskGroupReward: 2,
			TaskRewardItems: []data.TaskRewardItemVO{{TaskID: 1, RewardTemplateID: 2}},
		},
	}
	require.NoError(t, svc.EditVersion(context.Background(), id, 1, content))
	_, err = svc.PublishCampaign(context.Background(), id, "op")
	require.NoError(t, err)

	err = svc.EditVersion(context.Background(), id, 1, content)
	require.True(t, data.IsCampaignDraftNotEditable(err))
}

func TestCampaignAdmin_Publish_validation(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "X")
	require.NoError(t, err)
	require.NoError(t, svc.EditVersion(context.Background(), id, 1, data.CampaignVO{Name: "X"}))

	_, err = svc.PublishCampaign(context.Background(), id, "op")
	require.True(t, data.IsCampaignPublishInvalid(err))
}

func TestCampaignAdmin_Publish_alreadyPublishedLatest(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "X")
	require.NoError(t, err)
	require.NoError(t, svc.EditVersion(context.Background(), id, 1, data.CampaignVO{
		Name: "X", Budget: data.BudgetVO{ProjectID: 1, ProjectName: "p"},
		RewardRules: data.CampaignRewardRuleVO{
			TaskGroupID: 1, TaskGroupReward: 2,
			TaskRewardItems: []data.TaskRewardItemVO{{TaskID: 1, RewardTemplateID: 2}},
		},
	}))
	_, err = svc.PublishCampaign(context.Background(), id, "op")
	require.NoError(t, err)
	_, err = svc.PublishCampaign(context.Background(), id, "op")
	require.True(t, data.IsCampaignNoDraftToPublish(err))
}

func TestCampaignAdmin_ListCampaigns_paging(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	for i := 0; i < 3; i++ {
		_, err := svc.CreateCampaign(context.Background(), "C")
		require.NoError(t, err)
	}

	items, total, err := svc.ListCampaigns(data.CampaignListReq{Page: 2, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), items[0].ID)
	require.Equal(t, int64(1), items[0].Version)
}

func TestCampaignAdmin_ListCampaigns_defaultPaging(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	_, err := svc.CreateCampaign(context.Background(), "C")
	require.NoError(t, err)

	items, total, err := svc.ListCampaigns(data.CampaignListReq{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), items[0].Version)
}

func TestCampaignAdmin_ListCampaigns_statusFilter(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	_, err := svc.CreateCampaign(context.Background(), "C")
	require.NoError(t, err)
	published := model.CampaignStatusPublished

	items, total, err := svc.ListCampaigns(data.CampaignListReq{Status: &published})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, items)
}

func TestCampaignAdmin_ListCampaigns_maxVersion(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "C")
	require.NoError(t, err)
	v2, err := svc.CreateVersion(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, 2, v2)

	items, total, err := svc.ListCampaigns(data.CampaignListReq{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, int64(2), items[0].Version)
}

func TestCampaignAdmin_CreateVersion_copiesPreviousContent(t *testing.T) {
	svc, _, _, _ := newAdminSvc()
	id, err := svc.CreateCampaign(context.Background(), "X")
	require.NoError(t, err)
	require.NoError(t, svc.EditVersion(context.Background(), id, 1, data.CampaignVO{Name: "Copied", Market: "US"}))

	v2, err := svc.CreateVersion(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, 2, v2)
	detail, err := svc.GetCampaign(id)
	require.NoError(t, err)
	require.Equal(t, int64(2), detail.Version)
	require.Equal(t, "Copied", detail.Name)
	require.Equal(t, "US", detail.Market)
}
