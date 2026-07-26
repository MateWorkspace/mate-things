package domaincontractscache

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Firmware interface {
	GetById(ctx context.Context, id uuid.UUID) (firmware *domainmodels.Firmware, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, firmware *domainmodels.Firmware) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (firmware *domainmodels.Firmware, hit bool, err error)
	SetByName(ctx context.Context, name string, firmware *domainmodels.Firmware) error
	DeleteByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID) (pagination Pagination[domainmodels.Firmware], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, pagination Pagination[domainmodels.Firmware]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
