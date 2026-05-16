package service_test

import (
	"errors"
	"testing"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubUserAccountRepo struct {
	creditCalled bool
}

func (r *stubUserAccountRepo) GetByUserAndCurrency(int64, string) (*model.UserAccount, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *stubUserAccountRepo) CreditWithTransaction(txn *model.AccountTransaction) (float64, error) {
	r.creditCalled = true
	return txn.Amount, nil
}

type stubAccountTxnRepo struct{}

func (stubAccountTxnRepo) SumAmountByType(int64, string, string) (float64, error) {
	return 0, nil
}

func (stubAccountTxnRepo) List(mysql.AccountTransactionListFilter) ([]model.AccountTransaction, error) {
	return nil, nil
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
			require.True(t, service.IsInvalidAccountInput(err))
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
	require.False(t, service.IsInvalidAccountInput(err))
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
