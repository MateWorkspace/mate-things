package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeClass interface {
	Create(
		ctx context.Context,
		name string,
		description *string,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (nodeClass *domainmodels.NodeClass, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (nodeClass *domainmodels.NodeClass, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
	) (nodeClasses []domainmodels.NodeClass, total int, err error)

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
