package compositionseeder

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type driver struct {
	// Logger
	zlogLogger zerolog.Logger
	slogLogger *slog.Logger

	// Postgres
	postgresPool *pgxpool.Pool
	dt           pgxdt.Pgxdt
	tx           pgxdt.Transactor
}

func (l *launcher) newDriver(ctx context.Context) error {
	const tag = path + "/driver"

	var (
		zlogLogger zerolog.Logger
		slogLogger *slog.Logger
	)
	if config.LoggerFormat == domainmodels.LoggerFormatJson {
		zlogLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		slogLogger = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		))
	} else {
		zlogLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
		slogLogger = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		))
	}

	postgresPool, err := newPostgresPool(ctx)
	if err != nil {
		l.logError(
			slogLogger, zlogLogger,
			tag, "Failed to create PostgreSQL pool",
			domainmodels.LoggerMeta{"error": err.Error()},
		)
		return err
	}

	l.drv = &driver{
		zlogLogger:   zlogLogger,
		slogLogger:   slogLogger,
		postgresPool: postgresPool,
		dt:           pgxdt.NewPgxdt(postgresPool),
		tx:           pgxdt.NewTransactor(postgresPool),
	}

	l.logInfo(
		slogLogger, zlogLogger,
		tag, "Driver initialized", nil,
	)

	return nil
}

func (d *driver) cleanup() {
	if d.postgresPool != nil {
		d.postgresPool.Close()
	}
}

func newPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(config.PostgresUsername),
		url.QueryEscape(config.PostgresPassword),
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDatabase,
		config.PostgresSslMode,
	)

	pgxConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return nil, err
	}
	if config.PostgresMaxConnections > 0 {
		pgxConfig.MaxConns = int32(config.PostgresMaxConnections)
	}

	connectCtx, cancel := context.WithTimeout(ctx, config.PostgresConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, pgxConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func (l *launcher) logError(slogLogger *slog.Logger, zlogLogger zerolog.Logger, tag string, message string, meta domainmodels.LoggerMeta) {
	switch config.LoggerDriver {
	case "zerolog":
		zlogLogger.Error().
			Str("tag", tag).
			Any("meta", meta).
			Msg(message)
	default:
		slogLogger.Error(
			message,
			slog.String("tag", tag),
			slog.Any("meta", meta),
		)
	}
}

func (l *launcher) logInfo(slogLogger *slog.Logger, zlogLogger zerolog.Logger, tag string, message string, meta domainmodels.LoggerMeta) {
	switch config.LoggerDriver {
	case "zerolog":
		zlogLogger.Info().
			Str("tag", tag).
			Any("meta", meta).
			Msg(message)
	default:
		slogLogger.Info(
			message,
			slog.String("tag", tag),
			slog.Any("meta", meta),
		)
	}
}
