package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ircop/wgadmin/wglib/slave"
	"go.uber.org/zap"
)

func main() {
	remote := os.Getenv("REMOTE")
	ifname := os.Getenv("IFNAME")
	login := os.Getenv("LOGIN")
	password := os.Getenv("PASSWORD")
	saveTemplate := os.Getenv("SAVE_TEMPLATE")
	savePath := os.Getenv("SAVE_PATH")

	skipTLSVerify := false
	if os.Getenv("SKIPTLSVERIFY") == "true" {
		skipTLSVerify = true
	}

	if remote == "" || ifname == "" {
		usage()
		os.Exit(1)
	}

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("failed to init logger: %s\n", err.Error())
		os.Exit(1)
	}

	sd, err := slave.NewSLave(ifname, logger)
	if err != nil {
		logger.Error("faield to initialize slave daemon", zap.String("ifname", ifname), zap.Error(err))
		os.Exit(1)
	}

	if savePath == "" && saveTemplate == "" {
		logger.Info("skipping autosave setup: no SAVE_PATH or SAVE_TEMPLATE provided.")
	} else if saveTemplate, err = checkAutosave(saveTemplate, savePath); err != nil {
		logger.Error("autosave verification failed", zap.Error(err))
		os.Exit(1)
	}

	config := slave.Config{
		RemoteAddr:     remote,
		BasicAuthLogin: login,
		BasicAuthPW:    password,
		SaveTemplate:   saveTemplate,
		SavePath:       savePath,
		SkipTLSVerify:  skipTLSVerify,
	}
	if err = sd.Run(config); err != nil {
		logger.Error("failed to run slave daemon", zap.String("remote", remote), zap.Error(err))
		os.Exit(1)
	}
}

func checkAutosave(template string, path string) (string, error) {
	bts, err := ioutil.ReadFile(template)
	if err != nil {
		return "", err
	}

	templateContents := string(bts)

	if !strings.Contains(templateContents, "[Interface]") ||
		!strings.Contains(templateContents, "Address") ||
		!strings.Contains(templateContents, "PrivateKey") ||
		!strings.Contains(templateContents, "ListenPort") {
		return "", errors.New("wrong config template contents")
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if f != nil {
		f.Close()
	}

	if err != nil {
		return "", err
	}

	return templateContents, nil
}

func usage() {
	fmt.Printf("usage:\n\tREMOTE=domain.tld/endpoint IFNAME=wgintf %s\n", os.Args[0])
}
