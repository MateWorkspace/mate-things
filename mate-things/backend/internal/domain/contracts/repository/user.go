package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type User interface {
	Create(
		ctx context.Context,
		roleId uuid.UUID,
		name string,
		bio *string,
		username string,
		passwordHash string,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (user *domainmodels.User, err error)

	ReadByUsername(
		ctx context.Context,
		username string,
	) (user *domainmodels.User, err error)

	ReadPermissions(
		ctx context.Context,
		userId uuid.UUID,
	) (permissions []domainmodels.Permission, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		roleId *uuid.UUID,
	) (users []domainmodels.User, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		roleId *uuid.UUID,
		name *string,
		bio *string,
		username *string,
		passwordHash *string,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
