package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/common/middleware"
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
	r.Use(gin.Logger(), gin.Recovery())

	basicGroup := r.Group(serviceURIPrefix)
	basicGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	basicGroup.GET("/ping", otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware(), api.Ping)

	clientGroup := basicGroup.Group("/:client")
	clientGroup.Use(middleware.ValidateClient(), otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware())
	clientGroup.GET("/hello", api.SayHello)

	admin := basicGroup.Group("/admin")
	admin.Use(otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware())
	{
		admin.POST("/campaigns", api.AdminCreateCampaign)
		admin.PUT("/campaigns/:campaignId", api.AdminUpdateCampaign)
		admin.GET("/campaigns", api.AdminListCampaigns)
		admin.GET("/campaigns/:campaignId", api.AdminGetCampaign)
		admin.POST("/campaigns/:campaignId/publish", api.AdminPublishCampaign)

		admin.POST("/landing-pages", api.AdminCreateLandingPage)
		admin.PUT("/landing-pages/:landingPageId", api.AdminUpdateLandingPage)
		admin.GET("/landing-pages", api.AdminListLandingPages)
		admin.GET("/landing-pages/:landingPageId", api.AdminGetLandingPage)
		admin.POST("/landing-pages/:landingPageId/publish", api.AdminPublishLandingPage)
	}

	// User-facing campaign APIs (same :client as middleware; no extra path segment).
	// Correct URL: /campaign-center-api/v1/web/campaigns/:id/landing-page
	web := basicGroup.Group("/web")
	web.Use(otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware())
	{
		web.GET("/campaigns/:campaignId/landing-page", api.UserGetCampaignLanding)
		web.POST("/campaigns/:campaignId/join", api.UserJoinCampaign)
		web.POST("/campaigns/:campaignId/top-up", api.UserSimulateTopUp)
	}

	return r
}
