include .env

.PHONY: default
default: run

.PHONY: install
install:
	@go mod download && ./bin/install.sh && go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps

.PHONY: update
update:
	@go get -u ./... && go mod tidy

.PHONY: setup
setup:
	@$(CHROME_PATH) --remote-debugging-port=$(CDP_PORT) --profile-directory=Default

.PHONY: run
run:
	@go run ./cmd/webscraper

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
