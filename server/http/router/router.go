package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/auth"
	"github.com/lianjin/campaign-center-api/server/config"
	_ "github.com/lianjin/campaign-center-api/server/docs"
	"github.com/lianjin/campaign-center-api/server/http/api"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const serviceURIPrefix = "/campaign-center-api/v1"

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(log.RecoveryMiddleware())
	r.Use(corsMiddleware())

	basicGroup := r.Group(serviceURIPrefix)
	basicGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	admin := basicGroup.Group("/admin")
	admin.Use(otelgin.Middleware(data.ServiceName), log.HTTPObservabilityMiddleware(), auth.RequireAdmin())
	{
		admin.POST("/campaigns", api.AdminCreateCampaign)
		admin.PUT("/campaigns/:campaignId", api.AdminUpdateCampaign)
		admin.GET("/campaigns", api.AdminListCampaigns)
		admin.GET("/campaigns/:campaignId", api.AdminGetCampaign)
		admin.POST("/campaigns/:campaignId/publish", api.AdminPublishCampaign)
		admin.POST("/campaigns/:campaignId/archive", api.AdminArchiveCampaign)
		admin.GET("/campaigns/:campaignId/performance/summary", api.AdminGetCampaignPerformanceSummary)
		admin.GET("/campaigns/:campaignId/performance/daily", api.AdminListCampaignDailyPerformance)
		admin.GET("/campaigns/:campaignId/participations", api.AdminListCampaignParticipations)

		admin.POST("/landing-pages/:landingPageId/translations/generate", api.AdminGenerateLandingTranslation)
		admin.GET("/landing-pages/:landingPageId/translations", api.AdminListLandingPageTranslatedLangs)
		admin.PUT("/landing-pages/:landingPageId/translations/:lang", api.AdminPutLandingTranslation)
		admin.POST("/landing-pages", api.AdminCreateLandingPage)
		admin.PUT("/landing-pages/:landingPageId", api.AdminUpdateLandingPage)
		admin.GET("/landing-pages", api.AdminListLandingPages)
		admin.GET("/landing-pages/:landingPageId/detail/:lang", api.AdminGetLandingPageLocaleDetail)
		admin.GET("/landing-pages/:landingPageId", api.AdminGetLandingPage)
		admin.POST("/landing-pages/:landingPageId/publish", api.AdminPublishLandingPage)
	}

	// User-facing campaign APIs
	web := basicGroup.Group("/web")
	web.Use(otelgin.Middleware(data.ServiceName), log.HTTPObservabilityMiddleware(), auth.RequireUser())
	{
		web.GET("/account/summary", api.UserGetAccountSummary)
		web.GET("/account/transactions", api.UserListAccountTransactions)
		web.GET("/campaigns/:campaignId/landing-page", api.UserGetCampaignLanding)
		web.POST("/campaigns/:campaignId/join", api.UserJoinCampaign)
		web.POST("/campaigns/:campaignId/top-up", api.UserSimulateTopUp)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: config.Config.SystemConfig.AllowedOrigins,
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
