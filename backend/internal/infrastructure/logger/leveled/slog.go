package infrastructureloggerleveled

import (
	"context"
	"log/slog"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type slogImpl struct {
	logger      *slog.Logger
	loggerLevel domainmodels.LoggerLevel
}

func NewSlogImpl(
	logger *slog.Logger,
	loggerLevel domainmodels.LoggerLevel,
) domaincontractslogger.Leveled {
	return &slogImpl{
		logger:      logger,
		loggerLevel: loggerLevel,
	}
}

func (s *slogImpl) Error(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if s.loggerLevel.Order() < domainmodels.LoggerLevelError.Order() {
		return
	}
	attr := s.buildAttrs(tag, meta)
	s.logger.LogAttrs(ctx, slog.LevelError, message, attr...)
}

func (s *slogImpl) Warn(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if s.loggerLevel.Order() < domainmodels.LoggerLevelWarn.Order() {
		return
	}
	attr := s.buildAttrs(tag, meta)
	s.logger.LogAttrs(ctx, slog.LevelWarn, message, attr...)
}

func (s *slogImpl) Info(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if s.loggerLevel.Order() < domainmodels.LoggerLevelInfo.Order() {
		return
	}
	attr := s.buildAttrs(tag, meta)
	s.logger.LogAttrs(ctx, slog.LevelInfo, message, attr...)
}

func (s *slogImpl) Debug(ctx context.Context, tag string, message string, meta domainmodels.LoggerMeta) {
	if s.loggerLevel.Order() < domainmodels.LoggerLevelDebug.Order() {
		return
	}
	attr := s.buildAttrs(tag, meta)
	s.logger.LogAttrs(ctx, slog.LevelDebug, message, attr...)
}

func (s *slogImpl) buildAttrs(tag string, meta domainmodels.LoggerMeta) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("tag", tag),
		slog.Any("meta", normalizeMeta(meta)),
	}
	return attrs
}
