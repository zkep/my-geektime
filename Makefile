PROJECT:=mygeektime

.PHONY: build


all: build githook

build:
	go vet ./cmd/api
	go build -ldflags "-X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o .


githook:
	git config core.hooksPath .githooks
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	chmod 777 .githooks/commit-msg
	chmod 777 .githooks/pre-commit
	chmod 777 .githooks/pre-push

run: build
	gofmt -w ./
	mygeektime \
    --help