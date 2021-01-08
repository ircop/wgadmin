package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pnforge/wgadmin/wglib/master"
	"go.uber.org/zap"
)

func main() {
	login := os.Getenv("LOGIN")
	password := os.Getenv("PASSWORD")

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("failed to init logger: %s\n", err.Error())
		os.Exit(1)
	}

	// register websocket daemon http handler
	md := master.NewMasterDaemon(logger, time.Second*5)
	wsHandler, err := md.Run()

	if err != nil {
		logger.Error("failed to start daemon", zap.Error(err))
		os.Exit(1)
	}

	apiServer := &APIServer{
		Md:     md,
		logger: logger,
	}

	router := getRouter(wsHandler, apiServer, login, password)
	// start https listener
	if err = http.ListenAndServe(":80", router); err != nil {
		logger.Error("failed to start listener", zap.Error(err))
		os.Exit(1)
	}
}
