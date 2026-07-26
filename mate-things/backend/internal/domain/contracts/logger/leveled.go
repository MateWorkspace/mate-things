package domaincontractslogger

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type Leveled interface {
	Error(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta)
	Warn(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta)
	Info(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta)
	Debug(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta)
}
