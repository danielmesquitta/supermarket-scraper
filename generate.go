package root

//go:generate wire-config -c internal/config/wire/wire.go -o internal/app/webscraper/wire.go -m github.com/danielmesquitta/supermarket-web-scraper/internal/app/webscraper -e prod
