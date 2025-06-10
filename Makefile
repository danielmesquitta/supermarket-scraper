include .env

schema=./sql/schema.prisma

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
	@if [ ! -d tmp/user-data-dir ]; then cp -r $(USER_DATA_DIR) tmp/user-data-dir; fi && $(CHROME_PATH) --remote-debugging-port=$(CDP_PORT) --user-data-dir=tmp/user-data-dir

.PHONY: run
run:
	@go run ./cmd/run

.PHONY: retry
retry:
	@go run ./cmd/retry

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

.PHONY: migrate
migrate:
	@prisma-client-go migrate dev --schema=$(schema) --skip-generate

.PHONY: deploy_migrations
deploy_migrations:
	@prisma-client-go migrate deploy --schema=$(schema)

.PHONY: reset_db
reset_db:
	@prisma-client-go migrate reset --schema=$(schema) --skip-generate
