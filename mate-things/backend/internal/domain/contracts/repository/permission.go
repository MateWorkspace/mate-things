package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Permission interface {
	Create(
		ctx context.Context,
		name string,
		description *string,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (permission *domainmodels.Permission, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (permission *domainmodels.Permission, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
	) (permissions []domainmodels.Permission, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		name *string,
		description *string,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
