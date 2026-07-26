package infrastructureloggerleveled

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/rs/zerolog"
)

type zerologImpl struct {
	logger      zerolog.Logger
	loggerLevel domainmodels.LoggerLevel
}

func NewZerologImpl(
	logger zerolog.Logger,
	loggerLevel domainmodels.LoggerLevel,
) domaincontractslogger.Leveled {
	return &zerologImpl{
		logger:      logger,
		loggerLevel: loggerLevel,
	}
}

func (z *zerologImpl) Error(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if z.loggerLevel.Order() < domainmodels.LoggerLevelError.Order() {
		return
	}
	z.buildEvent(ctx, z.logger.Error(), tag, meta).Msg(message)
}

func (z *zerologImpl) Warn(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if z.loggerLevel.Order() < domainmodels.LoggerLevelWarn.Order() {
		return
	}
	z.buildEvent(ctx, z.logger.Warn(), tag, meta).Msg(message)
}

func (z *zerologImpl) Info(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if z.loggerLevel.Order() < domainmodels.LoggerLevelInfo.Order() {
		return
	}
	z.buildEvent(ctx, z.logger.Info(), tag, meta).Msg(message)
}

func (z *zerologImpl) Debug(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if z.loggerLevel.Order() < domainmodels.LoggerLevelDebug.Order() {
		return
	}
	z.buildEvent(ctx, z.logger.Debug(), tag, meta).Msg(message)
}

func (z *zerologImpl) buildEvent(ctx context.Context, event *zerolog.Event, tag string, meta domainmodels.LoggerMeta) *zerolog.Event {
	event = event.Ctx(ctx).
		Str("tag", tag).
		Any("meta", meta)
	return event
}
