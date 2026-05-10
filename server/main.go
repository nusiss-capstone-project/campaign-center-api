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
	"github.com/lianjin/campaign-center-api/server/repository"
	"github.com/lianjin/campaign-center-api/server/telemetry"
)

var sigCh = make(chan os.Signal, 1)

func main() {
	fmt.Println("Starting campaign-center-api...")
	config.Init()
	log.InitLogger()
	log.Logger.Info("Logger initialized.")
	if _, err := repository.Init(); err != nil {
		log.Logger.Warnw("MySQL initialization failed", "error", err)
	} else if config.Config.MySQLConfig != nil && config.Config.MySQLConfig.Enabled {
		log.Logger.Info("MySQL initialized.")
	} else {
		log.Logger.Info("MySQL initialization skipped.")
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
