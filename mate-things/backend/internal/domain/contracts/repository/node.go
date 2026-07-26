package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Node interface {
	Create(
		ctx context.Context,
		nodeClassId uuid.UUID,
		deviceId string,
		deviceInfo string,
		name string,
		firmwareName string,
		description *string,
		isConnected bool,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	UpsertRegistration(
		ctx context.Context,
		deviceId string,
		deviceInfo string,
		firmwareName string,
	) (node *domainmodels.Node, created bool, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (node *domainmodels.Node, err error)

	ReadByDeviceId(
		ctx context.Context,
		deviceId string,
	) (node *domainmodels.Node, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		nodeClassId *uuid.UUID,
		firmwareName *string,
	) (nodes []domainmodels.Node, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		nodeClassId *uuid.UUID,
		deviceId *string,
		deviceInfo *string,
		name *string,
		firmwareName *string,
		description *string,
		isConnected *bool,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
