package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/service"
)

// LandingPageLocaleDetailData documents StandardResponse.data: title/description/terms from translation when present, else default row; banner/status/timestamps from landing page.
type LandingPageLocaleDetailData struct {
	ID             int64  `json:"id" example:"2001"`
	Lang           string `json:"lang" example:"ja"`
	DefaultLang    string `json:"defaultLang" example:"en"`
	BannerImageURL string `json:"bannerImageUrl"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Terms          string `json:"terms"`
	Status         int16  `json:"status"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// LandingPageLocaleDetailHTTPResponse documents HTTP 200 for locale detail.
type LandingPageLocaleDetailHTTPResponse struct {
	Code    int                         `json:"code" example:"0"`
	Message string                      `json:"message" example:"success"`
	Data    LandingPageLocaleDetailData `json:"data"`
}

// LandingPageBody create/update landing page JSON body.
type LandingPageBody struct {
	DefaultLang    string `json:"defaultLang" binding:"required"`
	BannerImageURL string `json:"bannerImageUrl" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description" binding:"required"`
	Terms          string `json:"terms" binding:"required"`
}

// AdminCreateLandingPage creates a draft landing page.
// @Summary Create landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param body body LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages [post]
func AdminCreateLandingPage(c *gin.Context) {
	var req LandingPageBody
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageAdminService()
	id, status, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		DefaultLang:    req.DefaultLang,
		BannerImageURL: req.BannerImageURL,
		Title:          req.Title,
		Description:    req.Description,
		Terms:          req.Terms,
	})
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"landingPageId": id, "status": status})
}

// AdminUpdateLandingPage updates a draft landing page.
// @Summary Update landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse "success"
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
	var req LandingPageBody
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageAdminService()
	err = svc.UpdateDraftLandingPage(id, service.CreateLandingPageParams{
		DefaultLang:    req.DefaultLang,
		BannerImageURL: req.BannerImageURL,
		Title:          req.Title,
		Description:    req.Description,
		Terms:          req.Terms,
	})
	if err != nil {
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
	data.OK(c, gin.H{"landingPageId": id})
}

// AdminListLandingPages lists landing pages.
// @Summary List landing pages (admin)
// @Tags admin-landing-page
// @Produce json
// @Param page query int false "Page"
// @Param pageSize query int false "Page size"
// @Param status query int false "Status filter"
// @Param defaultLang query string false "Default language filter e.g. en"
// @Success 200 {object} data.StandardResponse "success"
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
	svc := service.GetLandingPageAdminService()
	items, total, err := svc.ListLandingPages(mysql.LandingPageListFilter{
		Page: page, PageSize: pageSize, Status: statusPtr, DefaultLang: defaultLang,
	})
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{
			"id":             it.ID,
			"defaultLang":    it.DefaultLang,
			"title":          it.Title,
			"status":         it.Status,
			"bannerImageUrl": it.BannerImageURL,
			"createdAt":      it.CreatedAt.Format(time.RFC3339),
		})
	}
	data.OK(c, gin.H{"total": total, "items": out})
}

// AdminGetLandingPage returns landing page detail with optional lang overlay.
// @Summary Get landing page (admin)
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang query string false "Requested language (falls back to default)"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId} [get]
func AdminGetLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	lang := c.Query("lang")
	svc := service.GetLandingPageAdminService()
	p, err := svc.GetLandingPage(id, lang)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, landingPageDetailPayload(p))
}

// AdminGetLandingPageLocaleDetail returns merged detail for one landing page and locale (path).
// @Summary Get landing page detail by locale (admin)
// @Description title/description/terms come from campaign_landing_page_translations when a row exists for lang; otherwise from campaign_landing_pages. bannerImageUrl, status, timestamps always from campaign_landing_pages.
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang path string true "Locale tag, e.g. ja, zh-CN"
// @Success 200 {object} LandingPageLocaleDetailHTTPResponse "success"
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
	svc := service.GetLandingPageAdminService()
	p, err := svc.GetLandingPage(id, lang)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, landingPageDetailPayload(p))
}

func landingPageDetailPayload(p *service.LandingPageDetailView) gin.H {
	return gin.H{
		"id":             p.ID,
		"lang":           p.Lang,
		"defaultLang":    p.DefaultLang,
		"bannerImageUrl": p.BannerImageURL,
		"title":          p.Title,
		"description":    p.Description,
		"terms":          p.Terms,
		"status":         p.Status,
		"createdAt":      p.CreatedAt.Format(time.RFC3339),
		"updatedAt":      p.UpdatedAt.Format(time.RFC3339),
	}
}

// AdminPublishLandingPage publishes a landing page.
// @Summary Publish landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId}/publish [post]
func AdminPublishLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	var req PublishOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageAdminService()
	updated, err := svc.PublishLandingPage(id, req.Operator)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"landingPageId": updated.ID, "status": updated.Status})
}
