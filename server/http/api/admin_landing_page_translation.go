package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/service"
)

// GenerateLandingTranslationReq is the JSON body for machine translation (no persist).
type GenerateLandingTranslationReq struct {
	SourceLang  string `json:"sourceLang" binding:"required" example:"en"`
	TargetLang  string `json:"targetLang" binding:"required" example:"ja"`
	Title       string `json:"title" example:"Top up <amount> to get reward"`
	Description string `json:"description" example:"Recharge now and receive <reward_amount> bonus"`
	Terms       string `json:"terms" example:"Reward will expire in <days> days"`
}

// PutLandingTranslationReq is the JSON body for upserting one locale.
type PutLandingTranslationReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Terms       string `json:"terms" binding:"required"`
	Operator    string `json:"operator" example:"admin"`
}

// GenerateLandingTranslationData is the shape of StandardResponse.data after generate.
type GenerateLandingTranslationData struct {
	Lang        string `json:"lang" example:"ja"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Terms       string `json:"terms"`
}

// GenerateLandingTranslationHTTPResponse documents the 200 envelope for Swagger.
type GenerateLandingTranslationHTTPResponse struct {
	Code    int                            `json:"code" example:"0"`
	Message string                         `json:"message" example:"success"`
	Data    GenerateLandingTranslationData `json:"data"`
}

// PutLandingTranslationData is the shape of StandardResponse.data after save.
type PutLandingTranslationData struct {
	LandingPageID int64  `json:"landingPageId"`
	Lang          string `json:"lang"`
}

// PutLandingTranslationHTTPResponse documents the 200 envelope after save.
type PutLandingTranslationHTTPResponse struct {
	Code    int                       `json:"code" example:"0"`
	Message string                    `json:"message" example:"success"`
	Data    PutLandingTranslationData `json:"data"`
}

// AdminGenerateLandingTranslation calls OpenAI to translate fields (preview only).
// @Summary Generate landing page translation preview (admin)
// @Description Returns LLM-translated title/description/terms for the given landing page. Does not persist.
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body GenerateLandingTranslationReq true "Source/target languages and optional source copy (falls back to landing page fields when empty)"
// @Success 200 {object} GenerateLandingTranslationHTTPResponse "success"
// @Failure 400 {object} data.StandardResponse "validation or empty source"
// @Failure 404 {object} data.StandardResponse "landing page not found"
// @Failure 503 {object} data.StandardResponse "OpenAI not configured or database unavailable"
// @Failure 500 {object} data.StandardResponse "internal error"
// @Router /admin/landing-pages/{landingPageId}/translations/generate [post]
func AdminGenerateLandingTranslation(c *gin.Context) {
	landingPageID, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	var req GenerateLandingTranslationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageTranslationService()
	out, err := svc.GenerateTranslation(c.Request.Context(), service.GenerateTranslationParams{
		LandingPageID: landingPageID,
		SourceLang:    req.SourceLang,
		TargetLang:    req.TargetLang,
		Title:         req.Title,
		Description:   req.Description,
		Terms:         req.Terms,
	})
	if err != nil {
		if service.IsOpenAINotConfigured(err) {
			data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
			return
		}
		if service.IsTranslationSourceEmpty(err) {
			data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
			return
		}
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{
		"lang": out.Lang, "title": out.Title,
		"description": out.Description, "terms": out.Terms,
	})
}

// AdminPutLandingTranslation upserts one translation for a landing page and locale.
// @Summary Upsert landing page translation (admin)
// @Description Creates or updates campaign_landing_page_translations for the given landing page and language code.
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang path string true "BCP-47 or short language tag, e.g. ja, zh-CN"
// @Param body body PutLandingTranslationReq true "Translated fields"
// @Success 200 {object} PutLandingTranslationHTTPResponse "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 404 {object} data.StandardResponse "landing page not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Failure 500 {object} data.StandardResponse "internal error"
// @Router /admin/landing-pages/{landingPageId}/translations/{lang} [put]
func AdminPutLandingTranslation(c *gin.Context) {
	landingPageID, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	lang := c.Param("lang")
	if lang == "" {
		data.JSON(c, http.StatusBadRequest, -1, "invalid lang", nil)
		return
	}
	var req PutLandingTranslationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetLandingPageTranslationService()
	err = svc.SaveTranslation(c.Request.Context(), service.SaveTranslationParams{
		LandingPageID: landingPageID,
		Lang:          lang,
		Title:         req.Title,
		Description:   req.Description,
		Terms:         req.Terms,
		Operator:      req.Operator,
	})
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"landingPageId": landingPageID, "lang": lang})
}

// LandingPageTranslatedLangsData is StandardResponse.data for persisted translation locales.
type LandingPageTranslatedLangsData struct {
	Langs []string `json:"langs"`
}

// LandingPageTranslatedLangsHTTPResponse documents HTTP 200 for translation lang list.
type LandingPageTranslatedLangsHTTPResponse struct {
	Code    int                            `json:"code" example:"0"`
	Message string                         `json:"message" example:"success"`
	Data    LandingPageTranslatedLangsData `json:"data"`
}

// AdminListLandingPageTranslatedLangs returns language codes that have rows in campaign_landing_page_translations.
// @Summary List translated locales for a landing page (admin)
// @Description Distinct lang values from the translation table only (excludes default_lang unless a translation row exists).
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Success 200 {object} LandingPageTranslatedLangsHTTPResponse "success"
// @Failure 400 {object} data.StandardResponse "invalid path"
// @Failure 404 {object} data.StandardResponse "landing page not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/landing-pages/{landingPageId}/translations [get]
func AdminListLandingPageTranslatedLangs(c *gin.Context) {
	landingPageID, err := strconv.ParseInt(c.Param("landingPageId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid landingPageId", nil)
		return
	}
	svc := service.GetLandingPageTranslationService()
	langs, err := svc.ListTranslatedLangs(c.Request.Context(), landingPageID)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"langs": langs})
}
