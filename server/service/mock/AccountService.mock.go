package mock

import (
	"github.com/lianjin/campaign-center-api/server/service"
	mock "github.com/stretchr/testify/mock"
)

// MockAccountService is a testify mock for AccountService.
type MockAccountService struct {
	mock.Mock
}

func (_m *MockAccountService) GetSummary(userID int64, currency string) (*service.AccountSummary, error) {
	ret := _m.Called(userID, currency)
	var r0 *service.AccountSummary
	if v := ret.Get(0); v != nil {
		r0 = v.(*service.AccountSummary)
	}
	return r0, ret.Error(1)
}

func (_m *MockAccountService) Recharge(userID int64, amount float64, currency string) (*service.RechargeResult, error) {
	ret := _m.Called(userID, amount, currency)
	var r0 *service.RechargeResult
	if v := ret.Get(0); v != nil {
		r0 = v.(*service.RechargeResult)
	}
	return r0, ret.Error(1)
}

func (_m *MockAccountService) CreditCampaignReward(userID, campaignID int64, amount float64, currency string) (*service.RechargeResult, error) {
	ret := _m.Called(userID, campaignID, amount, currency)
	var r0 *service.RechargeResult
	if v := ret.Get(0); v != nil {
		r0 = v.(*service.RechargeResult)
	}
	return r0, ret.Error(1)
}

func (_m *MockAccountService) ListTransactions(userID int64, txnType string, cursorID int64, limit int) ([]service.AccountTransactionItem, error) {
	ret := _m.Called(userID, txnType, cursorID, limit)
	var r0 []service.AccountTransactionItem
	if v := ret.Get(0); v != nil {
		r0 = v.([]service.AccountTransactionItem)
	}
	return r0, ret.Error(1)
}
