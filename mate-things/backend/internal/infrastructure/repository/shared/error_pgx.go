package infrastructurerepositoryshared

import (
	"errors"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func MapPgxError(message string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503", "23505":
			return domainmodels.NewError(message, domainmodels.ErrTypeConflict, err)
		case "22001", "22P02", "23502", "23514":
			return domainmodels.NewError(message, domainmodels.ErrTypeValidation, err)
		}
	}

	return domainmodels.NewError(message, domainmodels.ErrTypeUnknown, err)
}

func QueryBuildError(message string, err error) error {
	return domainmodels.NewError(message, domainmodels.ErrTypeFailure, err)
}

func NotFound(message string, err error) error {
	return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
}
