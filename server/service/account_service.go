package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// AccountSummary is the client account overview.
type AccountSummary struct {
	UserID                    int64   `json:"userId"`
	Currency                  string  `json:"currency"`
	Balance                   float64 `json:"balance"`
	TotalRechargeAmount       float64 `json:"totalRechargeAmount"`
	TotalCampaignRewardAmount float64 `json:"totalCampaignRewardAmount"`
}

// AccountTransactionItem is a single ledger row for clients.
type AccountTransactionItem struct {
	TransactionNo string  `json:"transactionNo"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	BalanceAfter  float64 `json:"balanceAfter"`
	CreatedAt     string  `json:"createdAt"`
}

// RechargeResult is returned after a successful recharge.
type RechargeResult struct {
	TransactionNo string  `json:"transactionNo"`
	BalanceAfter  float64 `json:"balanceAfter"`
}

// AccountService handles user account balance and ledger entries.
type AccountService interface {
	GetSummary(userID int64, currency string) (*AccountSummary, error)
	Recharge(userID int64, amount float64, currency string) (*RechargeResult, error)
	CreditCampaignReward(userID, campaignID int64, amount float64, currency string) (*RechargeResult, error)
	ListTransactions(userID int64, txnType string, cursorID int64, limit int) ([]AccountTransactionItem, error)
}

type accountService struct {
	accounts mysql.UserAccountRepository
	txns     mysql.AccountTransactionRepository
	users    mysql.UserRepository
}

var _ AccountService = (*accountService)(nil)

var (
	accountServiceOnce sync.Once
	accountServiceInst AccountService
)

// NewAccountService builds an account service (for tests).
func NewAccountService(
	accounts mysql.UserAccountRepository,
	txns mysql.AccountTransactionRepository,
	users mysql.UserRepository,
) AccountService {
	return &accountService{accounts: accounts, txns: txns, users: users}
}

// GetAccountService returns the singleton account service.
func GetAccountService() AccountService {
	accountServiceOnce.Do(func() {
		accountServiceInst = NewAccountService(
			mysql.GetUserAccountRepository(),
			mysql.GetAccountTransactionRepository(),
			mysql.GetUserRepository(),
		)
	})
	return accountServiceInst
}

func (s *accountService) GetSummary(userID int64, currency string) (*AccountSummary, error) {
	if currency == "" {
		currency = model.DefaultCurrency
	}
	balance := 0.0
	if acc, err := s.accounts.GetByUserAndCurrency(userID, currency); err == nil {
		balance = acc.Balance
	} else if !mysql.IsNotFound(err) {
		return nil, err
	}
	recharge, err := s.txns.SumAmountByType(userID, currency, model.AccountTxnTypeRecharge)
	if err != nil {
		return nil, err
	}
	reward, err := s.txns.SumAmountByType(userID, currency, model.AccountTxnTypeCampaignReward)
	if err != nil {
		return nil, err
	}
	return &AccountSummary{
		UserID: userID, Currency: currency, Balance: balance,
		TotalRechargeAmount: recharge, TotalCampaignRewardAmount: reward,
	}, nil
}

func (s *accountService) Recharge(userID int64, amount float64, currency string) (*RechargeResult, error) {
	if err := s.validateRechargeInput(userID, amount); err != nil {
		return nil, err
	}
	return s.credit(userID, amount, currency, model.AccountTxnTypeRecharge,
		model.AccountTxnRelatedTypeRecharge, 0, "recharge")
}

func (s *accountService) validateRechargeInput(userID int64, amount float64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: userID must be positive", data.ErrInvalidAccountInput)
	}
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", data.ErrInvalidAccountInput)
	}
	if _, err := s.users.GetByID(userID); err != nil {
		if mysql.IsNotFound(err) {
			return fmt.Errorf("%w: user not found", data.ErrInvalidAccountInput)
		}
		return err
	}
	return nil
}

func (s *accountService) CreditCampaignReward(
	userID, campaignID int64, amount float64, currency string,
) (*RechargeResult, error) {
	return s.credit(userID, amount, currency, model.AccountTxnTypeCampaignReward,
		model.AccountTxnRelatedTypeCampaign, campaignID, "campaign reward")
}

func (s *accountService) ListTransactions(
	userID int64, txnType string, cursorID int64, limit int,
) ([]AccountTransactionItem, error) {
	rows, err := s.txns.List(mysql.AccountTransactionListFilter{
		UserID: userID, Type: txnType, CursorID: cursorID, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]AccountTransactionItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, AccountTransactionItem{
			TransactionNo: row.TransactionNo,
			Type:          row.Type,
			Amount:        row.Amount,
			Currency:      row.Currency,
			Status:        row.Status,
			BalanceAfter:  row.BalanceAfter,
			CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *accountService) credit(
	userID int64, amount float64, currency, txnType, relatedType string,
	relatedID int64, remark string,
) (*RechargeResult, error) {
	if currency == "" {
		currency = model.DefaultCurrency
	}
	txn := &model.AccountTransaction{
		TransactionNo: newAccountTransactionNo(),
		UserID:        userID,
		Currency:      currency,
		Amount:        amount,
		Type:          txnType,
		Status:        model.AccountTxnStatusSuccess,
		RelatedType:   relatedType,
		RelatedID:     relatedID,
		Remark:        remark,
		UpdatedAt:     time.Now(),
	}
	balanceAfter, err := s.accounts.CreditWithTransaction(txn)
	if err != nil {
		return nil, err
	}
	return &RechargeResult{TransactionNo: txn.TransactionNo, BalanceAfter: balanceAfter}, nil
}

func newAccountTransactionNo() string {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return fmt.Sprintf("TXN%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("TXN%d%s", time.Now().UnixNano(), hex.EncodeToString(suffix[:]))
}
