PROJECT:=mygeektime


.PHONY: build

build:
	git config core.hooksPath .githooks
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	chmod 777 .githooks/pre-commit
	chmod 777 .githooks/pre-push
