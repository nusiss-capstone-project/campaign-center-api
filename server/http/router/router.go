package router

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	_ "github.com/nusiss-capstone-project/campaign-center-api/server/docs"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/api"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	commonauth "github.com/nusiss-capstone-project/identity-mservice/common/auth"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const serviceURIPrefix = "/campaign-center-api/v1"

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(log.RecoveryMiddleware())
	r.Use(otelgin.Middleware(data.ServiceName))
	r.Use(log.HTTPResponseIDMiddleware())
	r.Use(corsMiddleware())

	basicGroup := r.Group(serviceURIPrefix)
	{
		// High-frequency / non-business routes: no HTTP access log.
		basicGroup.GET("/ping", api.Ping)
		basicGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Business routes: enable request access logging.
		apiGroup := basicGroup.Group("")
		apiGroup.Use(commonauth.AuditMiddleware(func(ctx context.Context) commonauth.AuditLogger {
			return log.WithContext(ctx)
		}))
		apiGroup.Use(log.HTTPObservabilityMiddleware())
		{
			admin := apiGroup.Group("/admin")
			admin.Use(commonauth.RequireRole([]string{commonauth.RoleCampaignOps, commonauth.RoleAdmin}))
			{
				admin.POST("/campaigns", api.AdminCreateCampaign)
				admin.POST("/campaigns/:campaignId/versions", api.AdminCreateCampaignVersion)
				admin.PUT("/campaigns/:campaignId/versions/:version", api.AdminEditCampaignVersion)
				admin.GET("/campaigns", api.AdminListCampaigns)
				admin.GET("/campaigns/:campaignId", api.AdminGetCampaign)
				admin.POST("/campaigns/:campaignId/publish", api.AdminPublishCampaign)
				admin.GET("/campaigns/:campaignId/users/:userId", api.AdminGetCampaignUser)
				admin.GET("/campaigns/:campaignId/users", api.AdminListCampaignUsers)

				admin.POST("/landing-pages/:landingPageId/translations/generate", api.AdminGenerateLandingTranslation)
				admin.GET("/landing-pages/:landingPageId/translations", api.AdminListLandingPageTranslatedLangs)
				admin.PUT("/landing-pages/:landingPageId/translations/:lang", api.AdminPutLandingTranslation)
				admin.POST("/landing-pages", api.AdminCreateLandingPage)
				admin.PUT("/landing-pages/:landingPageId", api.AdminUpdateLandingPage)
				admin.GET("/landing-pages", api.AdminListLandingPages)
				admin.GET("/landing-pages/:landingPageId/detail/:lang", api.AdminGetLandingPageLocaleDetail)
				admin.GET("/landing-pages/:landingPageId", api.AdminGetLandingPage)
				admin.POST("/landing-pages/:landingPageId/publish", api.AdminPublishLandingPage)
				admin.POST("/images/upload", api.AdminUploadImage)
			}

			web := apiGroup.Group("/web")
			web.Use(commonauth.RequireUser())
			{
				web.GET("/campaigns", api.UserListCampaigns)
				web.GET("/campaigns/:campaignId/landing-page", api.UserGetCampaignLanding)
				web.GET("/campaigns/:campaignId/rules", api.UserGetCampaignRules)
				web.POST("/campaigns/:campaignId/join", api.UserJoinCampaign)
			}
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins(),
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			commonauth.HeaderInternalUserID, commonauth.HeaderUserRole,
			log.RequestIDHeader, log.TraceIDHeader,
		},
		ExposeHeaders: []string{
			"Content-Length", commonauth.HeaderInternalUserID, commonauth.HeaderUserRole,
			log.RequestIDHeader, log.TraceIDHeader,
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func allowedOrigins() []string {
	if config.Config == nil || config.Config.SystemConfig == nil {
		return []string{}
	}
	return config.Config.SystemConfig.AllowedOrigins
}
