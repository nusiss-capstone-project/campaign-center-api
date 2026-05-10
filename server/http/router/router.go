package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/api"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const serviceURIPrefix = "/campaign-center-api/v1"

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	basicGroup := r.Group(serviceURIPrefix)
	basicGroup.GET("/ping", otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware(), api.Ping)

	clientGroup := basicGroup.Group("/:client")
	clientGroup.Use(otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware())
	clientGroup.GET("/hello", api.SayHello)

	return r
}
