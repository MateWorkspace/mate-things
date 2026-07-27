package compositionmain

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttevent "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/event"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type driver struct {
	// Logger
	zlogLogger zerolog.Logger
	slogLogger *slog.Logger

	// HTTP
	echo            *echo.Echo
	echoStartConfig echo.StartConfig

	// Postgres
	postgresPool *pgxpool.Pool
	dt           pgxdt.Pgxdt
	tx           pgxdt.Transactor

	// Redis
	redisClient redis.UniversalClient

	// MinIO
	minioClient *minio.Client

	// MQTT
	mqttClient mqtt.Client
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

	echoInstance := echo.New()
	echoStartConfig := echo.StartConfig{
		Address:         config.HttpServerAddress,
		GracefulTimeout: config.HttpServerShutdownTimeout,
		BeforeServeFunc: func(server *http.Server) error {
			server.ReadTimeout = config.HttpServerReadTimeout
			return nil
		},
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

	redisClient, err := newRedisClient(ctx)
	if err != nil {
		l.logError(
			slogLogger, zlogLogger,
			tag, "Failed to create Redis client",
			domainmodels.LoggerMeta{"error": err.Error()},
		)
		postgresPool.Close()
		return err
	}

	minioClient, err := newMinioClient(ctx)
	if err != nil {
		l.logError(
			slogLogger, zlogLogger,
			tag, "Failed to create MinIO client",
			domainmodels.LoggerMeta{"error": err.Error()},
		)
		_ = redisClient.Close()
		postgresPool.Close()
		return err
	}

	mqttClient, err := newMqttClient()
	if err != nil {
		l.logError(
			slogLogger, zlogLogger,
			tag, "Failed to connect MQTT client",
			domainmodels.LoggerMeta{"error": err.Error()},
		)
		_ = redisClient.Close()
		postgresPool.Close()
		return err
	}

	l.drv = &driver{
		zlogLogger:      zlogLogger,
		slogLogger:      slogLogger,
		echo:            echoInstance,
		echoStartConfig: echoStartConfig,
		postgresPool:    postgresPool,
		dt:              pgxdt.NewPgxdt(postgresPool),
		tx:              pgxdt.NewTransactor(postgresPool),
		redisClient:     redisClient,
		minioClient:     minioClient,
		mqttClient:      mqttClient,
	}

	l.logInfo(
		slogLogger, zlogLogger,
		tag, "Driver initialized", nil,
	)

	return nil
}

func (d *driver) cleanup() {
	if d.mqttClient != nil && d.mqttClient.IsConnected() {
		d.mqttClient.Disconnect(250)
	}
	if d.redisClient != nil {
		_ = d.redisClient.Close()
	}
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

func newRedisClient(ctx context.Context) (redis.UniversalClient, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:       config.RedisAddresses,
		Password:    config.RedisPassword,
		DB:          config.RedisDatabase,
		DialTimeout: config.RedisConnectTimeout,
	})

	connectCtx, cancel := context.WithTimeout(ctx, config.RedisConnectTimeout)
	defer cancel()

	if err := client.Ping(connectCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func newMinioClient(ctx context.Context) (*minio.Client, error) {
	client, err := minio.New(config.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinioAccessKey, config.MinioSecretKey, ""),
		Secure: config.MinioUseSsl,
	})
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, config.MinioConnectTimeout)
	defer cancel()

	exists, err := client.BucketExists(connectCtx, config.MinioBucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(connectCtx, config.MinioBucket, minio.MakeBucketOptions{
			Region: config.MinioRegion,
		}); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func newMqttClient() (mqtt.Client, error) {
	mqttOptions := mqtt.NewClientOptions().
		AddBroker(config.MqttBrokerUrl).
		SetClientID(config.MqttClientId).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectTimeout(config.MqttConnectTimeout).
		SetKeepAlive(config.MqttKeepAlive).
		SetPingTimeout(config.MqttPingTimeout).
		SetOnConnectHandler(presentationmqttevent.OnConnect).
		SetConnectionLostHandler(presentationmqttevent.OnDisconnected).
		SetDefaultPublishHandler(presentationmqttevent.OnMessage)
	if config.MqttUsername != "" {
		mqttOptions.SetUsername(config.MqttUsername)
	}
	if config.MqttPassword != "" {
		mqttOptions.SetPassword(config.MqttPassword)
	}

	mqttClient := mqtt.NewClient(mqttOptions)
	mqttToken := mqttClient.Connect()
	if !mqttToken.WaitTimeout(config.MqttConnectTimeout) {
		return nil, domainmodels.NewError("failed to connect MQTT client", domainmodels.ErrTypeTimeout, nil)
	}
	if err := mqttToken.Error(); err != nil {
		return nil, err
	}

	return mqttClient, nil
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
