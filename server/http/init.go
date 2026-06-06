package http

import (
	"fmt"
	"os"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/router"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
)

func Init(exitSig chan os.Signal) {
	r := router.NewRouter()
	log.Logger.Infof("Campaign Center API HTTP server starting...")
	if err := r.Run(fmt.Sprintf("%s:%d", config.Config.HttpConfig.Host, config.Config.HttpConfig.Port)); err != nil {
		log.Logger.Errorf("Failed to run HTTP server: %v", err)
		if exitSig != nil {
			exitSig <- os.Interrupt
		}
	}
}
