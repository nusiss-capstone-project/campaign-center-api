package service

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/nusiss-capstone-project/campaign-center-api/server/kafka"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type rewardRulesStub struct {
	byRef map[string][]model.CampaignRewardRule
}

func (r rewardRulesStub) ListByCampaignID(int64) ([]model.CampaignRewardRule, error) { return nil, nil }
func (r rewardRulesStub) ListByRef(refClient string, refID int64) ([]model.CampaignRewardRule, error) {
	return r.byRef[refClient+":"+strconv.FormatInt(refID, 10)], nil
}
func (r rewardRulesStub) ReplaceByCampaignID(context.Context, int64, []model.CampaignRewardRule) error {
	return nil
}
func (r rewardRulesStub) DeleteByCampaignID(context.Context, int64) error { return nil }

type ledgerStub struct {
	byKey             map[string]*model.CampaignUserRewardLedger
	byID              map[int64]*model.CampaignUserRewardLedger
	nextID            int64
	updates           []string
	createErr         error
	afterCreateLookup map[string]*model.CampaignUserRewardLedger
	createFailedOnce  bool
}

func ledgerKey(userID, campaignID, ruleID int64) string {
	return strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(campaignID, 10) + ":" + strconv.FormatInt(ruleID, 10)
}

func (l *ledgerStub) Create(_ context.Context, row *model.CampaignUserRewardLedger) error {
	if l.createErr != nil {
		l.createFailedOnce = true
		return l.createErr
	}
	l.nextID++
	row.ID = l.nextID
	cp := *row
	l.byID[row.ID] = &cp
	l.byKey[ledgerKey(row.UserID, row.CampaignID, row.RuleID)] = &cp
	return nil
}
func (l *ledgerStub) GetByID(id int64) (*model.CampaignUserRewardLedger, error) {
	return l.byID[id], nil
}
func (l *ledgerStub) GetByUserCampaignRule(userID, campaignID, ruleID int64) (*model.CampaignUserRewardLedger, error) {
	key := ledgerKey(userID, campaignID, ruleID)
	if l.createFailedOnce && l.afterCreateLookup != nil {
		if row, ok := l.afterCreateLookup[key]; ok {
			return row, nil
		}
	}
	return l.byKey[key], nil
}
func (l *ledgerStub) UpdateStatusAndVoucher(_ context.Context, id int64, status, voucherID string) error {
	row := l.byID[id]
	if row == nil {
		return nil
	}
	row.RewardStatus = status
	if voucherID != "" {
		row.VoucherID = voucherID
	}
	l.updates = append(l.updates, status)
	return nil
}
func (l *ledgerStub) ListByCampaignAndUser(int64, int64) ([]model.CampaignUserRewardLedger, error) {
	return nil, nil
}

type rewardClientStub struct {
	voucher string
	err     error
	calls   int
}

func (r *rewardClientStub) Distribute(context.Context, proxy.RewardDistributeRequest) (*proxy.RewardDistributeResult, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	return &proxy.RewardDistributeResult{VoucherID: r.voucher}, nil
}

func TestCampaignRewardEventService_HandleTaskCompleted_createsLedgerAndRewards(t *testing.T) {
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{},
		byID:  map[int64]*model.CampaignUserRewardLedger{},
	}
	reward := &rewardClientStub{voucher: "voucher-1"}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99, Status: model.CampaignStatusPublished}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task:7": {{ID: 11, CampaignID: 1, RefClient: "task", RefID: 7, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	err := svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed", GroupID: 1, CompletedTaskCount: 1, TotalTaskCount: 2,
	})
	require.NoError(t, err)
	require.Equal(t, 1, reward.calls)
	require.Equal(t, model.LedgerRewardStatusDistributing, ledgers.byID[1].RewardStatus)
	require.Equal(t, "voucher-1", ledgers.byID[1].VoucherID)
}

func TestCampaignRewardEventService_HandleTaskCompleted_skipsWhenVoucherExists(t *testing.T) {
	existing := &model.CampaignUserRewardLedger{
		ID: 3, UserID: 5, CampaignID: 1, RuleID: 11,
		RewardStatus: model.LedgerRewardStatusDistributing, VoucherID: "v-already",
	}
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{"5:1:11": existing},
		byID:  map[int64]*model.CampaignUserRewardLedger{3: existing},
	}
	reward := &rewardClientStub{voucher: "x"}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task:7": {{ID: 11, CampaignID: 1, RefClient: "task", RefID: 7, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed", GroupID: 1, CompletedTaskCount: 1, TotalTaskCount: 1,
	}))
	require.Equal(t, 0, reward.calls)
}

func TestCampaignRewardEventService_HandleTaskCompleted_retriesWhenNoVoucher(t *testing.T) {
	existing := &model.CampaignUserRewardLedger{
		ID: 3, UserID: 5, CampaignID: 1, RuleID: 11,
		RewardStatus: model.LedgerRewardStatusDistributeFail, VoucherID: "",
	}
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{"5:1:11": existing},
		byID:  map[int64]*model.CampaignUserRewardLedger{3: existing},
	}
	reward := &rewardClientStub{voucher: "v-retry"}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task:7": {{ID: 11, CampaignID: 1, RefClient: "task", RefID: 7, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed", GroupID: 1, CompletedTaskCount: 1, TotalTaskCount: 1,
	}))
	require.Equal(t, 1, reward.calls)
	require.Equal(t, model.LedgerRewardStatusDistributing, ledgers.byID[3].RewardStatus)
	require.Equal(t, "v-retry", ledgers.byID[3].VoucherID)
}

func TestCampaignRewardEventService_HandleRewardResult(t *testing.T) {
	row := &model.CampaignUserRewardLedger{ID: 9, RewardStatus: model.LedgerRewardStatusDistributing, VoucherID: "v0"}
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{},
		byID:  map[int64]*model.CampaignUserRewardLedger{9: row},
	}
	svc := NewCampaignRewardEventService(nil, rewardRulesStub{}, &webParticipantRepoStub{}, ledgers, &rewardClientStub{})
	require.NoError(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{
		ClientRefID: "9", VoucherID: "v1", Status: "DISTRIBUTED",
	}))
	require.Equal(t, model.LedgerRewardStatusDistributeSuccess, ledgers.byID[9].RewardStatus)
	require.Equal(t, "v1", ledgers.byID[9].VoucherID)
}

func TestCampaignRewardEventService_HandleTaskCompleted_ignoredAndNoRules(t *testing.T) {
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{}},
		&webParticipantRepoStub{},
		&ledgerStub{byKey: map[string]*model.CampaignUserRewardLedger{}, byID: map[int64]*model.CampaignUserRewardLedger{}},
		&rewardClientStub{},
	)
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 1, UserID: 1, Status: "pending",
	}))
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 1, UserID: 1, Status: "completed", GroupID: 9, CompletedTaskCount: 1, TotalTaskCount: 2,
	}))
}

func TestCampaignRewardEventService_HandleTaskCompleted_groupRuleAndNotJoined(t *testing.T) {
	reward := &rewardClientStub{voucher: "g1"}
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{},
		byID:  map[int64]*model.CampaignUserRewardLedger{},
	}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task_group:3": {{ID: 21, CampaignID: 1, RefClient: "task_group", RefID: 3, RewardTemplateID: 88}},
			"task:7":       {{ID: 22, CampaignID: 2, RefClient: "task", RefID: 7, RewardTemplateID: 0}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed", GroupID: 3, CompletedTaskCount: 2, TotalTaskCount: 2,
	}))
	require.Equal(t, 1, reward.calls)
	require.Equal(t, "g1", ledgers.byID[1].VoucherID)

	// Same group rule but user not joined → skip distribute.
	reward2 := &rewardClientStub{voucher: "x"}
	svc2 := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task_group:3": {{ID: 21, CampaignID: 1, RefClient: "task_group", RefID: 3, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: nil},
		&ledgerStub{byKey: map[string]*model.CampaignUserRewardLedger{}, byID: map[int64]*model.CampaignUserRewardLedger{}},
		reward2,
	)
	require.NoError(t, svc2.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 0, UserID: 5, Status: "completed", GroupID: 3, CompletedTaskCount: 1, TotalTaskCount: 1,
	}))
	require.Equal(t, 0, reward2.calls)
}

func TestCampaignRewardEventService_HandleTaskCompleted_createConflictRetries(t *testing.T) {
	existing := &model.CampaignUserRewardLedger{
		ID: 8, UserID: 5, CampaignID: 1, RuleID: 11,
		RewardStatus: model.LedgerRewardStatusPendingDistribution, VoucherID: "",
	}
	ledgers := &ledgerStub{
		byKey:     map[string]*model.CampaignUserRewardLedger{},
		byID:      map[int64]*model.CampaignUserRewardLedger{8: existing},
		createErr: errors.New("duplicate"),
		afterCreateLookup: map[string]*model.CampaignUserRewardLedger{
			"5:1:11": existing,
		},
	}
	reward := &rewardClientStub{voucher: "v-conflict"}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task:7": {{ID: 11, CampaignID: 1, RefClient: "task", RefID: 7, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	require.NoError(t, svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed",
	}))
	require.Equal(t, 1, reward.calls)
	require.Equal(t, "v-conflict", ledgers.byID[8].VoucherID)
}

func TestCampaignRewardEventService_HandleTaskCompleted_distributeFail(t *testing.T) {
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{},
		byID:  map[int64]*model.CampaignUserRewardLedger{},
	}
	reward := &rewardClientStub{err: errors.New("reward down")}
	svc := NewCampaignRewardEventService(
		&webCampaignRepoStub{campaigns: []model.Campaign{{ID: 1, BudgetProjectID: 99}}},
		rewardRulesStub{byRef: map[string][]model.CampaignRewardRule{
			"task:7": {{ID: 11, CampaignID: 1, RefClient: "task", RefID: 7, RewardTemplateID: 88}},
		}},
		&webParticipantRepoStub{row: &model.CampaignParticipant{CampaignID: 1, UserID: 5}},
		ledgers,
		reward,
	)
	err := svc.HandleTaskCompleted(context.Background(), TaskCompletedEvent{
		TaskID: 7, UserID: 5, Status: "completed",
	})
	require.Error(t, err)
	require.Equal(t, model.LedgerRewardStatusDistributeFail, ledgers.byID[1].RewardStatus)
}

func TestCampaignRewardEventService_HandleRewardResult_branches(t *testing.T) {
	row := &model.CampaignUserRewardLedger{ID: 9, RewardStatus: model.LedgerRewardStatusDistributing, VoucherID: "keep-me"}
	ledgers := &ledgerStub{
		byKey: map[string]*model.CampaignUserRewardLedger{},
		byID:  map[int64]*model.CampaignUserRewardLedger{9: row},
	}
	svc := NewCampaignRewardEventService(nil, rewardRulesStub{}, &webParticipantRepoStub{}, ledgers, &rewardClientStub{})

	require.Error(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{ClientRefID: ""}))
	require.Error(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{ClientRefID: "abc"}))
	require.ErrorIs(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{ClientRefID: "99"}), gorm.ErrRecordNotFound)

	require.NoError(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{
		ClientRefID: "9", Status: "FAILED",
	}))
	require.Equal(t, model.LedgerRewardStatusDistributeFail, ledgers.byID[9].RewardStatus)
	require.Equal(t, "keep-me", ledgers.byID[9].VoucherID)

	require.NoError(t, svc.HandleRewardResult(context.Background(), RewardDistributionResultEvent{
		ClientRefID: "9", Status: "WEIRD", VoucherID: "",
	}))
	require.Equal(t, model.LedgerRewardStatusDistributeFail, ledgers.byID[9].RewardStatus)
}

func TestParseRewardEvents(t *testing.T) {
	task, err := ParseTaskCompletedEvent(&kafka.Message{Value: []byte(`{"task_id":1,"user_id":2,"status":"completed"}`)})
	require.NoError(t, err)
	require.Equal(t, 1, task.TaskID)
	require.Equal(t, 2, task.UserID)

	_, err = ParseTaskCompletedEvent(&kafka.Message{Value: []byte(`{`)})
	require.Error(t, err)

	res, err := ParseRewardDistributionResultEvent(&kafka.Message{Value: []byte(`{"client_ref_id":"9","status":"FAILED"}`)})
	require.NoError(t, err)
	require.Equal(t, "9", res.ClientRefID)
	_, err = ParseRewardDistributionResultEvent(&kafka.Message{Value: []byte(`{`)})
	require.Error(t, err)
}
