package domaincontractscache

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type User interface {
	GetById(ctx context.Context, id uuid.UUID) (user *domainmodels.User, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, user *domainmodels.User) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByUsername(ctx context.Context, username string) (user *domainmodels.User, hit bool, err error)
	SetByUsername(ctx context.Context, username string, user *domainmodels.User) error
	DeleteByUsername(ctx context.Context, username string) error
	GetPermissions(ctx context.Context, userId uuid.UUID) (permissions []domainmodels.Permission, hit bool, err error)
	SetPermissions(ctx context.Context, userId uuid.UUID, permissions []domainmodels.Permission) error
	DeletePermissions(ctx context.Context, userId uuid.UUID) error
	InvalidatePermissions(ctx context.Context) error
	GetPagination(ctx context.Context, page int, limit int, search *string, roleId *uuid.UUID) (pagination Pagination[domainmodels.User], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, roleId *uuid.UUID, pagination Pagination[domainmodels.User]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
