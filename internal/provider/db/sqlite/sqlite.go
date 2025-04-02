package sqlite

import (
	"context"
	"log"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/provider/db/sqlite/schema"
)

type DB struct {
	db  *sqlx.DB
	gdb *goqu.Database
}

func New(
	e *env.Env,
) *DB {
	db, err := sqlx.Open("sqlite3", e.SQLiteDBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	gdb := goqu.New("sqlite3", db.DB)

	return &DB{
		db:  db,
		gdb: gdb,
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
		end := min(i+batchSize, len(products))

		batch := products[i:end]
		var records []goqu.Record

		for i := range batch {
			batch[i].ID = uuid.New().String()
			records = append(records, goqu.Record{
				schema.Product.ColumnID():    batch[i].ID,
				schema.Product.ColumnName():  batch[i].Name,
				schema.Product.ColumnPrice(): batch[i].Price,
			})
		}

		ds := d.gdb.Insert(schema.Product.Table()).Rows(records)
		sql, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return errs.New(err)
		}

		if _, err := d.db.ExecContext(ctx, sql, args...); err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func (d *DB) CreateError(
	ctx context.Context,
	errRec entity.Error,
) error {
	errRec.ID = uuid.New().String()

	ds := d.gdb.Insert(schema.Error.Table()).Rows(goqu.Record{
		schema.Error.ColumnID():         errRec.ID,
		schema.Error.ColumnMessage():    errRec.Message,
		schema.Error.ColumnType():       errRec.Type,
		schema.Error.ColumnStackTrace(): errRec.StackTrace,
		schema.Error.ColumnMetadata():   errRec.Metadata,
	})
	sql, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return errs.New(err)
	}

	if _, err := d.db.ExecContext(ctx, sql, args...); err != nil {
		return errs.New(err)
	}

	return nil
}

func (d *DB) ListProductProcessingErrors(
	ctx context.Context,
) ([]entity.Error, error) {
	ds := d.gdb.
		From(schema.Error.Table()).
		Select(schema.Error.ColumnAll()).
		Where(
			goqu.Ex{
				schema.Error.ColumnDeletedAt(): nil,
				schema.Error.ColumnType():      errs.ErrTypeFailedProcessingProductsPage,
			},
		).
		Order(goqu.C(schema.Error.ColumnCreatedAt()).Desc())

	sql, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, errs.New(err)
	}

	var errors []entity.Error
	if err := d.db.SelectContext(ctx, &errors, sql, args...); err != nil {
		return nil, errs.New(err)
	}

	return errors, nil
}

func (d *DB) DeleteErrors(
	ctx context.Context,
	ids []string,
) error {
	ds := d.gdb.
		Update(schema.Error.Table()).
		Set(goqu.Record{schema.Error.ColumnDeletedAt(): time.Now()}).
		Where(goqu.Ex{schema.Error.ColumnID(): ids})

	sql, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return errs.New(err)
	}

	_, err = d.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return errs.New(err)
	}

	return nil
}

func (d *DB) ListProductsByNames(
	ctx context.Context,
	names []string,
) ([]entity.Product, error) {
	ds := d.gdb.
		From(schema.Product.Table()).
		Select(schema.Product.ColumnAll()).
		Where(
			goqu.Ex{
				schema.Product.ColumnDeletedAt(): nil,
				schema.Product.ColumnName():      names,
			},
		)

	sql, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, errs.New(err)
	}

	var products []entity.Product
	if err := d.db.SelectContext(ctx, &products, sql, args...); err != nil {
		return nil, errs.New(err)
	}

	return products, nil
}
