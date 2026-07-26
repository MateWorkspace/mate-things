package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Role interface {
	Create(
		ctx context.Context,
		name string,
		description *string,
		isDefault *bool,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (role *domainmodels.Role, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (role *domainmodels.Role, err error)

	ReadDefault(
		ctx context.Context,
	) (role *domainmodels.Role, err error)

	ReadPermissions(
		ctx context.Context,
		roleId uuid.UUID,
	) (permissions []domainmodels.Permission, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
	) (roles []domainmodels.Role, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		name *string,
		description *string,
		isDefault *bool,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
