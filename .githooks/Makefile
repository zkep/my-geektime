PROJECT:=mygeektime


.PHONY: build

build:
	git config core.hooksPath .
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	chmod 777 pre-commit
	chmod 777 pre-push
