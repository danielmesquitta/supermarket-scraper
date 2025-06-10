package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/danielmesquitta/supermarket-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/db/sqlite"
)

type SaveErrorUseCase struct {
	db *sqlite.DB
}

func NewSaveErrorUseCase(db *sqlite.DB) *SaveErrorUseCase {
	return &SaveErrorUseCase{db: db}
}

func (u *SaveErrorUseCase) Execute(
	ctx context.Context,
	err error,
	metadata map[string]any,
) error {
	if err == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entityErr := entity.Error{
		Message: err.Error(),
		Type:    string(errs.ErrTypeUnknown),
	}

	var appErr *errs.Err
	if errors.As(err, &appErr) {
		if metadata != nil {
			metadataBytes, err := json.Marshal(metadata)
			if err != nil {
				return errs.New(err)
			}
			entityErr.Metadata = string(metadataBytes)
		}

		if appErr.Type != "" {
			entityErr.Type = string(appErr.Type)
		}

		entityErr.StackTrace = appErr.StackTrace
	}

	if err := u.db.CreateError(ctx, entityErr); err != nil {
		return errs.New(err)
	}

	return nil
}
