package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
)

// UserGetAccountSummary returns account overview for a user.
// @Summary Get account summary (user)
// @Tags user-account
// @Produce json
// @Param userId query int true "User ID"
// @Param currency query string false "Currency (default USDT)"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/account/summary [get]
func UserGetAccountSummary(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid userId", nil)
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
// @Param userId query int true "User ID"
// @Param type query string false "Transaction type RECHARGE or CAMPAIGN_REWARD"
// @Param cursor query int false "Pagination cursor (transaction id)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/account/transactions [get]
func UserListAccountTransactions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid userId", nil)
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
