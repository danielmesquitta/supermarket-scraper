include .env

.PHONY: default
default: run

.PHONY: install
install:
	@go mod download && ./bin/install.sh

.PHONY: update
update:
	@go get -u ./... && go mod tidy

.PHONY: run
run:
	@air -c .air.toml

.PHONY: clear
clear:
	@find ./tmp -mindepth 1 ! -name '.gitkeep' -delete

.PHONY: generate
generate:
	@go generate ./...

.PHONY: build
build:
	@GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o ./tmp/webscraper ./cmd/webscraper

.PHONY: lint
lint:
	@golangci-lint run && golines **/*.go -m 80 --dry-run

.PHONY: lint-fix
lint-fix:
	@golangci-lint run --fix && golines **/*.go -w -m 80
