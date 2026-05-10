package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/service"
)

// LandingPageBody create/update landing page JSON body.
type LandingPageBody struct {
	Language       string `json:"language" binding:"required"`
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
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param body body LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/admin/landing-pages [post]
func AdminCreateLandingPage(c *gin.Context) {
	var req LandingPageBody
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageAdminService()
	id, status, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		Language:       req.Language,
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
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param landingPageId path int true "Landing page ID"
// @Param body body LandingPageBody true "Landing page content"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 409 {object} data.StandardResponse "not draft"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/admin/landing-pages/{landingPageId} [put]
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
		Language:       req.Language,
		BannerImageURL: req.BannerImageURL,
		Title:          req.Title,
		Description:    req.Description,
		Terms:          req.Terms,
	})
	if err != nil {
		if service.IsLandingPageNotDraft(err) {
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
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param page query int false "Page"
// @Param pageSize query int false "Page size"
// @Param status query int false "Status filter"
// @Param language query string false "Language filter e.g. en-US"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/admin/landing-pages [get]
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
	language := c.Query("language")
	svc := service.GetLandingPageAdminService()
	items, total, err := svc.ListLandingPages(mysql.LandingPageListFilter{
		Page: page, PageSize: pageSize, Status: statusPtr, Language: language,
	})
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{
			"id":             it.ID,
			"language":       it.Language,
			"title":          it.Title,
			"status":         it.Status,
			"bannerImageUrl": it.BannerImageURL,
			"createdAt":      it.CreatedAt.Format(time.RFC3339),
		})
	}
	data.OK(c, gin.H{"total": total, "items": out})
}

// AdminGetLandingPage returns landing page detail.
// @Summary Get landing page (admin)
// @Tags admin-landing-page
// @Produce json
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param landingPageId path int true "Landing page ID"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/admin/landing-pages/{landingPageId} [get]
func AdminGetLandingPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	svc := service.GetLandingPageAdminService()
	p, err := svc.GetLandingPage(id)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{
		"id":             p.ID,
		"language":       p.Language,
		"bannerImageUrl": p.BannerImageURL,
		"title":          p.Title,
		"description":    p.Description,
		"terms":          p.Terms,
		"status":         p.Status,
	})
}

// AdminPublishLandingPage publishes a landing page.
// @Summary Publish landing page (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param landingPageId path int true "Landing page ID"
// @Param body body PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/admin/landing-pages/{landingPageId}/publish [post]
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
