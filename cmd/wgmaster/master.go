package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ircop/wgadmin/wglib/master"
	"go.uber.org/zap"
)

func main() {
	host := os.Getenv("HOST")
	portStr := os.Getenv("PORT")
	certPath := os.Getenv("CERT")
	keyPath := os.Getenv("KEY")

	login := os.Getenv("LOGIN")
	password := os.Getenv("PASSWORD")

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("failed to init logger: %s\n", err.Error())
		os.Exit(1)
	}

	port, err := strconv.Atoi(portStr)
	if host == "" || port < 0 || port > 65535 || certPath == "" || keyPath == "" || err != nil {
		usage()
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
	if err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, port), certPath, keyPath, router); err != nil {
		logger.Error("failed to start https listener", zap.Error(err))
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("usage:\n\tHOST=x.x.x.x PORT=yyyy CERT=/path/to/ssl.crt KEY=/path/to/ssl.key %s\n", os.Args[0])
}
