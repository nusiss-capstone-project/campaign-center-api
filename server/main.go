package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/lianjin/campaign-center-api/server/config"
	serverhttp "github.com/lianjin/campaign-center-api/server/http"
	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/telemetry"
)

var sigCh = make(chan os.Signal, 1)

// @title Campaign Center API
// @version 1.0
// @description HTTP API for Phase 1 user top-up campaigns (admin + user-facing). Operates under `/campaign-center-api/v1`; path segment `{client}` must be `merchant` or `customer`.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /campaign-center-api/v1
// @schemes http https
func main() {
	fmt.Println("Starting campaign-center-api...")
	config.Init()
	log.InitLogger()
	log.Logger.Info("Logger initialized.")
	if _, err := mysql.Init(); err != nil {
		log.Logger.Warnw("MySQL initialization failed", "error", err)
		panic(err)
	}
	shutdownTrace := telemetry.InitTracer()
	defer shutdownTrace()
	shutdownMetrics := telemetry.InitMetrics()
	defer shutdownMetrics()
	log.Logger.Info("Telemetry initialized.")
	go serverhttp.Init(sigCh)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	debug.PrintStack()
	log.Logger.Infof("Received signal: %v, shutting down", sig)
}
