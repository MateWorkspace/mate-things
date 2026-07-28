package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Role interface {
	GetById(ctx context.Context, id uuid.UUID) (role *domainmodels.Role, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, role *domainmodels.Role) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (role *domainmodels.Role, hit bool, err error)
	SetByName(ctx context.Context, name string, role *domainmodels.Role) error
	DeleteByName(ctx context.Context, name string) error
	GetDefault(ctx context.Context) (role *domainmodels.Role, hit bool, err error)
	SetDefault(ctx context.Context, role *domainmodels.Role) error
	DeleteDefault(ctx context.Context) error
	GetPermissions(ctx context.Context, roleId uuid.UUID) (permissions []domainmodels.Permission, hit bool, err error)
	SetPermissions(ctx context.Context, roleId uuid.UUID, permissions []domainmodels.Permission) error
	DeletePermissions(ctx context.Context, roleId uuid.UUID) error
	InvalidatePermissions(ctx context.Context) error
	GetPagination(ctx context.Context, page int, limit int, search *string) (pagination Pagination[domainmodels.Role], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, pagination Pagination[domainmodels.Role]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
