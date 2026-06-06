package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubUserAccountRepo struct {
	creditCalled bool
	account      *model.UserAccount
	getErr       error
}

func (r *stubUserAccountRepo) GetByUserAndCurrency(int64, string) (*model.UserAccount, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.account != nil {
		return r.account, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *stubUserAccountRepo) CreditWithTransaction(txn *model.AccountTransaction) (float64, error) {
	r.creditCalled = true
	return txn.Amount, nil
}

type stubAccountTxnRepo struct {
	rechargeSum float64
	rewardSum   float64
	listRows    []model.AccountTransaction
	listErr     error
}

func (r stubAccountTxnRepo) SumAmountByType(_ int64, _ string, txnType string) (float64, error) {
	switch txnType {
	case model.AccountTxnTypeRecharge:
		return r.rechargeSum, nil
	case model.AccountTxnTypeCampaignReward:
		return r.rewardSum, nil
	default:
		return 0, nil
	}
}

func (r stubAccountTxnRepo) List(mysql.AccountTransactionListFilter) ([]model.AccountTransaction, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listRows, nil
}

type stubUserRepo struct {
	err error
}

func (r stubUserRepo) GetByID(id int64) (*model.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &model.User{ID: id}, nil
}

func TestAccountService_Recharge_rejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name    string
		userID  int64
		amount  float64
		userErr error
	}{
		{name: "invalid user id", userID: 0, amount: 10},
		{name: "zero amount", userID: 1, amount: 0},
		{name: "negative amount", userID: 1, amount: -10},
		{name: "missing user", userID: 1, amount: 10, userErr: gorm.ErrRecordNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts := &stubUserAccountRepo{}
			svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{err: tc.userErr})

			_, err := svc.Recharge(tc.userID, tc.amount, model.DefaultCurrency)

			require.Error(t, err)
			require.True(t, data.IsInvalidAccountInput(err))
			require.False(t, accounts.creditCalled)
		})
	}
}

func TestAccountService_Recharge_propagatesUserLookupError(t *testing.T) {
	lookupErr := errors.New("db down")
	accounts := &stubUserAccountRepo{}
	svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{err: lookupErr})

	_, err := svc.Recharge(1, 10, model.DefaultCurrency)

	require.ErrorIs(t, err, lookupErr)
	require.False(t, data.IsInvalidAccountInput(err))
	require.False(t, accounts.creditCalled)
}

func TestAccountService_Recharge_validInputCreditsAccount(t *testing.T) {
	accounts := &stubUserAccountRepo{}
	svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{})

	out, err := svc.Recharge(1, 10, model.DefaultCurrency)

	require.NoError(t, err)
	require.True(t, accounts.creditCalled)
	require.Equal(t, float64(10), out.BalanceAfter)
}

func TestAccountService_GetSummary_withAccount(t *testing.T) {
	accounts := &stubUserAccountRepo{
		account: &model.UserAccount{UserID: 1, Currency: model.DefaultCurrency, Balance: 50},
	}
	txns := stubAccountTxnRepo{rechargeSum: 100, rewardSum: 20}
	svc := service.NewAccountService(accounts, txns, stubUserRepo{})

	out, err := svc.GetSummary(1, model.DefaultCurrency)

	require.NoError(t, err)
	require.Equal(t, float64(50), out.Balance)
	require.Equal(t, float64(100), out.TotalRechargeAmount)
	require.Equal(t, float64(20), out.TotalCampaignRewardAmount)
}

func TestAccountService_GetSummary_noAccountUsesZeroBalance(t *testing.T) {
	svc := service.NewAccountService(&stubUserAccountRepo{}, stubAccountTxnRepo{}, stubUserRepo{})

	out, err := svc.GetSummary(1, "")

	require.NoError(t, err)
	require.Equal(t, model.DefaultCurrency, out.Currency)
	require.Equal(t, float64(0), out.Balance)
}

func TestAccountService_GetSummary_propagatesAccountLookupError(t *testing.T) {
	lookupErr := errors.New("db down")
	accounts := &stubUserAccountRepo{getErr: lookupErr}
	svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{})

	_, err := svc.GetSummary(1, model.DefaultCurrency)

	require.ErrorIs(t, err, lookupErr)
}

func TestAccountService_ListTransactions_mapsRows(t *testing.T) {
	created := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	txns := stubAccountTxnRepo{listRows: []model.AccountTransaction{{
		TransactionNo: "TXN1", Type: model.AccountTxnTypeRecharge,
		Amount: 10, Currency: model.DefaultCurrency, Status: model.AccountTxnStatusSuccess,
		BalanceAfter: 110, CreatedAt: created,
	}}}
	svc := service.NewAccountService(&stubUserAccountRepo{}, txns, stubUserRepo{})

	out, err := svc.ListTransactions(1, model.AccountTxnTypeRecharge, 0, 20)

	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "TXN1", out[0].TransactionNo)
	require.Equal(t, created.Format(time.RFC3339), out[0].CreatedAt)
}

func TestAccountService_CreditCampaignReward_creditsAccount(t *testing.T) {
	accounts := &stubUserAccountRepo{}
	svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{})

	out, err := svc.CreditCampaignReward(1, 9, 15, model.DefaultCurrency)

	require.NoError(t, err)
	require.True(t, accounts.creditCalled)
	require.Equal(t, float64(15), out.BalanceAfter)
}

func TestAccountService_CreditCampaignReward_usesDefaultCurrency(t *testing.T) {
	accounts := &stubUserAccountRepo{}
	svc := service.NewAccountService(accounts, stubAccountTxnRepo{}, stubUserRepo{})

	_, err := svc.CreditCampaignReward(1, 9, 5, "")

	require.NoError(t, err)
	require.True(t, accounts.creditCalled)
}
