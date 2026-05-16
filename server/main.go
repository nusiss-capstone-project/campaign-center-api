package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/lianjin/campaign-center-api/server/config"
	serverhttp "github.com/lianjin/campaign-center-api/server/http"
	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/telemetry"
)

var sigCh = make(chan os.Signal, 1)

// @title Campaign Center API
// @version 1.0
// @description HTTP API for Phase 1 user top-up campaigns (admin + user-facing). Operates under `/campaign-center-api/v1`;
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
		log.Logger.Fatal("MySQL initialization failed", "error", err)
		panic("MySQL initialization failed")
	}
	shutdownTelemetry := telemetry.Init(context.Background())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(ctx); err != nil {
			log.Logger.Errorw("Telemetry shutdown failed", "error", err)
		}
	}()
	log.Logger.Info("Telemetry initialized.")
	go serverhttp.Init(sigCh)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	debug.PrintStack()
	log.Logger.Infof("Received signal: %v, shutting down", sig)
}
