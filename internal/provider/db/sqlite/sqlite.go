package sqlite

import (
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
)

type DB struct {
	db *sqlx.DB
}

func New(
	e *env.Env,
) *DB {
	db, err := sqlx.Open("sqlite3", e.SQLiteDBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	return &DB{
		db: db,
	}
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) CreateProducts(
	ctx context.Context,
	products []entity.Product,
) error {
	const batchSize = 500

	if len(products) == 0 {
		return nil
	}

	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]

		query := "INSERT INTO products (id, name, price) VALUES "
		var placeholders []string
		var args []any

		for j := range batch {
			batch[j].ID = uuid.New().String()

			placeholders = append(placeholders, "(?, ?, ?)")
			args = append(
				args,
				batch[j].ID,
				batch[j].Name,
				batch[j].Price,
			)
		}

		fullQuery := query + strings.Join(placeholders, ",")
		if _, err := d.db.ExecContext(ctx, fullQuery, args...); err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func (d *DB) CreateError(
	ctx context.Context,
	err entity.Error,
) error {
	query := `
    INSERT INTO errors (id, message, type, stack_trace, metadata)
    VALUES (:id, :message, :type, :stack_trace, :metadata)
  `

	err.ID = uuid.New().String()

	if _, err := d.db.NamedExecContext(ctx, query, err); err != nil {
		return errs.New(err)
	}

	return nil
}

func (d *DB) ListProductProcessingErrors(
	ctx context.Context,
) ([]entity.Error, error) {
	query := `
    SELECT * FROM errors
    WHERE deleted_at IS NULL
    AND type = 'failed_processing_products_page'
    ORDER BY created_at DESC
  `
	var errors []entity.Error
	if err := d.db.SelectContext(ctx, &errors, query); err != nil {
		return nil, errs.New(err)
	}

	return errors, nil
}

func (d *DB) DeleteErrors(
	ctx context.Context,
	ids []string,
) error {
	query := `
    UPDATE errors
    SET deleted_at = ?
    WHERE id IN (?)
  `

	query, args, err := sqlx.In(query, time.Now(), ids)
	if err != nil {
		return errs.New(err)
	}

	query = d.db.Rebind(query)

	_, err = d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errs.New(err)
	}

	return nil
}

func (d *DB) ListProductsByNames(
	ctx context.Context,
	names []string,
) ([]entity.Product, error) {
	query := `
    SELECT * FROM products
    WHERE deleted_at IS NULL
    AND name IN (?)
  `
	query, args, err := sqlx.In(query, names)
	if err != nil {
		return nil, errs.New(err)
	}

	query = d.db.Rebind(query)
	var products []entity.Product
	if err := d.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, errs.New(err)
	}

	return products, nil
}
