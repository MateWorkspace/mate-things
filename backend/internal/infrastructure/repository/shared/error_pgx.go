package infrastructurerepositoryshared

import (
	"errors"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ConflictMatch resolves a Postgres unique-violation to a specific
// ErrorType when the fired constraint's name contains Contains. Used at
// Create/Update call sites where a single statement can violate more than
// one unique constraint on the same table, so the generic 23505 code alone
// doesn't say which one fired.
type ConflictMatch struct {
	Contains string
	Type     domainmodels.ErrorType
}

func MapPgxError(message string, err error, conflicts ...ConflictMatch) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			for _, m := range conflicts {
				if strings.Contains(pgErr.ConstraintName, m.Contains) {
					return domainmodels.NewError(message, m.Type, err)
				}
			}
			return domainmodels.NewError(message, domainmodels.ErrTypeConflict, err)
		case "23503":
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
