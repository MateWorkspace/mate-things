package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeConfigValue interface {
	ReadByNodeId(
		ctx context.Context,
		nodeId uuid.UUID,
	) (values []domainmodels.NodeConfigValue, err error)

	ReadByNodeIdAndKey(
		ctx context.Context,
		nodeId uuid.UUID,
		key string,
	) (value *domainmodels.NodeConfigValue, err error)

	Upsert(
		ctx context.Context,
		nodeId uuid.UUID,
		firmwareId uuid.UUID,
		key string,
		value string,
		actorId *uuid.UUID,
	) (err error)
}
