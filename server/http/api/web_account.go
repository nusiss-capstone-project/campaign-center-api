package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/auth"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// UserGetAccountSummary returns account overview for a user.
// @Summary Get account summary (user)
// @Tags user-account
// @Produce json
// @Param currency query string false "Currency (default USDT)"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/account/summary [get]
func UserGetAccountSummary(c *gin.Context) {
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	currency := c.DefaultQuery("currency", model.DefaultCurrency)
	summary, err := service.GetAccountService().GetSummary(userID, currency)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.OK(c, summary)
}

// UserListAccountTransactions lists ledger entries for a user.
// @Summary List account transactions (user)
// @Tags user-account
// @Produce json
// @Param type query string false "Transaction type RECHARGE or CAMPAIGN_REWARD"
// @Param cursor query int false "Pagination cursor (transaction id)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/account/transactions [get]
func UserListAccountTransactions(c *gin.Context) {
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := service.GetAccountService().ListTransactions(userID, c.Query("type"), cursor, limit)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.OK(c, items)
}
