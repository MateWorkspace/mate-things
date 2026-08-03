package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeClassAction interface {
	Create(
		ctx context.Context,
		nodeClassId uuid.UUID,
		actionId uuid.UUID,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error)

	ReadByNodeClassIdAndActionId(
		ctx context.Context,
		nodeClassId uuid.UUID,
		actionId uuid.UUID,
	) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		nodeClassId *uuid.UUID,
		actionId *uuid.UUID,
	) (nodeClassActions []domainmodels.NodeClassAction, nodeClasses []domainmodels.NodeClass, actions []domainmodels.Action, total int, err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) (err error)

	DeleteByNodeClassIdAndActionId(
		ctx context.Context,
		nodeClassId *uuid.UUID,
		actionId *uuid.UUID,
	) (err error)
}
