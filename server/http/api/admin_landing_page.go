package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminCreateLandingPage creates a draft landing page.
// @Summary Create landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param body body data.LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageCreateResp} "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages [post]
func AdminCreateLandingPage(c *gin.Context) {
	var req data.LandingPageBody
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	resp, err := service.GetLandingPageAdminService().CreateLandingPage(req)
	if err != nil {
		if isLandingPageValidationErr(err) {
			data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

// AdminUpdateLandingPage updates a draft landing page.
// @Summary Update landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body data.LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageUpdateResp} "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 409 {object} data.StandardResponse "not draft"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId} [put]
func AdminUpdateLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	var req data.LandingPageBody
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	resp, err := service.GetLandingPageAdminService().UpdateDraftLandingPage(id, req)
	if err != nil {
		if isLandingPageValidationErr(err) {
			data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
			return
		}
		if data.IsLandingPageNotDraft(err) {
			data.JSON(c, http.StatusConflict, -1, err.Error(), nil)
			return
		}
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

// AdminListLandingPages lists landing pages.
// @Summary List landing pages (admin)
// @Tags admin-landing-page
// @Produce json
// @Param page query int false "Page"
// @Param pageSize query int false "Page size"
// @Param status query int false "Status filter"
// @Param defaultLang query string false "Default language filter e.g. en"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageListData} "success"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages [get]
func AdminListLandingPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var statusPtr *int16
	if s := c.Query("status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 16)
		if err == nil {
			x := int16(v)
			statusPtr = &x
		}
	}
	defaultLang := c.Query("defaultLang")
	resp, err := service.GetLandingPageAdminService().ListLandingPages(mysql.LandingPageListFilter{
		Page: page, PageSize: pageSize, Status: statusPtr, DefaultLang: defaultLang,
	})
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

// AdminGetLandingPage returns landing page detail with optional lang overlay.
// @Summary Get landing page (admin)
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang query string false "Requested language (falls back to default)"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageDetailVO} "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId} [get]
func AdminGetLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	resp, err := service.GetLandingPageAdminService().GetLandingPage(id, c.Query("lang"))
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

// AdminGetLandingPageLocaleDetail returns merged detail for one landing page and locale (path).
// @Summary Get landing page detail by locale (admin)
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang path string true "Locale tag, e.g. ja, zh-CN"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageDetailVO} "success"
// @Failure 400 {object} data.StandardResponse "invalid path"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId}/detail/{lang} [get]
func AdminGetLandingPageLocaleDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	lang := strings.TrimSpace(c.Param("lang"))
	if lang == "" {
		data.JSON(c, http.StatusBadRequest, -1, "invalid lang", nil)
		return
	}
	resp, err := service.GetLandingPageAdminService().GetLandingPage(id, lang)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

// AdminPublishLandingPage publishes a landing page.
// @Summary Publish landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body data.PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse{data=data.LandingPagePublishResp} "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId}/publish [post]
func AdminPublishLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	var req data.PublishOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	resp, err := service.GetLandingPageAdminService().PublishLandingPage(id, req.Operator)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, resp)
}

func isLandingPageValidationErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "steps") || strings.Contains(msg, "faq")
}
