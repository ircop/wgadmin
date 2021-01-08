GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

fmt:
	@echo "+ $@"
	goimports -w ${GOFILES}

build.master:
	@echo "+ $@"
	go build -o ./bin/app ./cmd/wgmaster/

app.dependencies.download:
	@echo "+ $@"
	GO111MODULE=on GOPRIVATE="github.com/pnforge,github.com/ircop" go mod download
