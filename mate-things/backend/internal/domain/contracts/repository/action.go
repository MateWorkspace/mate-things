package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Action interface {
	Create(
		ctx context.Context,
		nodeClassId uuid.UUID,
		name string,
		description *string,
		payloadSchemaName string,
		payloadSchemaVersion int32,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (action *domainmodels.Action, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (action *domainmodels.Action, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		nodeClassId *uuid.UUID,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
	) (actions []domainmodels.Action, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		nodeClassId *uuid.UUID,
		name *string,
		description *string,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
