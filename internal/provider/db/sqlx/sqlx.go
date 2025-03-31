package sqlx

import (
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"log"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/jmoiron/sqlx"
)

func New(
	e *env.Env,
) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", e.SQLiteDBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	return db
}
