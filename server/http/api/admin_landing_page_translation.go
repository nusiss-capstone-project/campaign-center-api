package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminGenerateLandingTranslation calls OpenAI to translate fields (preview only).
// @Summary Generate landing page translation preview (admin)
// @Description Returns LLM-translated title/description/terms/steps/faq. Does not persist.
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param body body data.GenerateLandingTranslationReq true "Source/target languages and optional source copy"
// @Success 200 {object} data.StandardResponse{data=data.GenerateLandingTranslationData} "success"
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
	var req data.GenerateLandingTranslationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	out, err := service.GetLandingPageTranslationService().GenerateTranslation(c.Request.Context(), service.GenerateTranslationParams{
		LandingPageID: landingPageID,
		SourceLang:    req.SourceLang,
		TargetLang:    req.TargetLang,
		Title:         req.Title,
		Description:   req.Description,
		Terms:         req.Terms,
		Steps:         req.Steps,
		Faq:           req.Faq,
	})
	if err != nil {
		if errors.Is(err, proxy.ErrOpenAINotConfigured) {
			data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
			return
		}
		if data.IsTranslationSourceEmpty(err) {
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
	data.OK(c, out)
}

// AdminPutLandingTranslation upserts one translation for a landing page and locale.
// @Summary Upsert landing page translation (admin)
// @Tags admin-landing-page
// @Accept json
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Param lang path string true "BCP-47 or short language tag, e.g. ja, zh-CN"
// @Param body body data.PutLandingTranslationReq true "Translated fields"
// @Success 200 {object} data.StandardResponse{data=data.PutLandingTranslationData} "success"
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
	var req data.PutLandingTranslationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	out, err := service.GetLandingPageTranslationService().SaveTranslation(c.Request.Context(), service.SaveTranslationParams{
		LandingPageID: landingPageID,
		Lang:          lang,
		Title:         req.Title,
		Description:   req.Description,
		Terms:         req.Terms,
		Steps:         req.Steps,
		Faq:           req.Faq,
		Operator:      req.Operator,
	})
	if err != nil {
		if isLandingPageValidationErr(err) {
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
	data.OK(c, out)
}

// AdminListLandingPageTranslatedLangs returns language codes that have translation rows.
// @Summary List translated locales for a landing page (admin)
// @Tags admin-landing-page
// @Produce json
// @Param landingPageId path int true "Landing page ID"
// @Success 200 {object} data.StandardResponse{data=data.LandingPageTranslatedLangsData} "success"
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
	out, err := service.GetLandingPageTranslationService().ListTranslatedLangs(c.Request.Context(), landingPageID)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "landing page not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, out)
}
