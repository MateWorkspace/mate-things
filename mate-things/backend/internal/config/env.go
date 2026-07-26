package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

var (
	LoggerFormat domainmodels.LoggerFormat = domainmodels.LoggerFormatJson
	LoggerLevel  domainmodels.LoggerLevel  = domainmodels.LoggerLevelInfo
	LoggerDriver string                    = "zerolog"

	HttpServerPort            int           = 8080
	HttpServerReadTimeout     time.Duration = 30 * time.Second
	HttpServerShutdownTimeout time.Duration = 10 * time.Second
	HttpCorsAllowedOrigins    []string      = []string{"*"}

	PostgresDatabaseUrl    string        = "postgres://postgres:postgres@127.0.0.1:5432/nusapala_things?sslmode=disable"
	PostgresMaxConnections int           = 20
	PostgresConnectTimeout time.Duration = 10 * time.Second

	RedisAddresses      []string      = []string{"127.0.0.1:6379"}
	RedisUsername       string        = ""
	RedisPassword       string        = ""
	RedisDatabase       int           = 0
	RedisConnectTimeout time.Duration = 10 * time.Second
	RedisCacheNamespace string        = "nusapala-things"
	RedisTtlIdentity    time.Duration = 30 * time.Minute
	RedisTtlPagination  time.Duration = 5 * time.Minute
	RedisTtlRelation    time.Duration = 10 * time.Minute

	MinioEndpoint       string        = "127.0.0.1:8333"
	MinioAccessKey      string        = "nusapala"
	MinioSecretKey      string        = "nusapala-secret"
	MinioBucket         string        = "nusapala-things"
	MinioUseSsl         bool          = false
	MinioRegion         string        = ""
	MinioConnectTimeout time.Duration = 10 * time.Second

	MqttBrokerUrl      string        = "tcp://127.0.0.1:1883"
	MqttClientId       string        = "nusapala-things-backend"
	MqttUsername       string        = ""
	MqttPassword       string        = ""
	MqttConnectTimeout time.Duration = 10 * time.Second
	MqttKeepAlive      time.Duration = 30 * time.Second
	MqttPingTimeout    time.Duration = 10 * time.Second

	TokenAccessSecret    string        = "nusapala-things-access-secret"
	TokenRefreshSecret   string        = "nusapala-things-refresh-secret"
	TokenAccessDuration  time.Duration = 15 * time.Minute
	TokenRefreshDuration time.Duration = 24 * time.Hour

	PasswordBcryptCost int = 10
)

func LoadEnv() {
	LoggerFormat = envGetLoggerFormat("LOGGER_FORMAT", LoggerFormat)
	LoggerLevel = envGetLoggerLevel("LOGGER_LEVEL", LoggerLevel)
	LoggerDriver = envGetString("LOGGER_DRIVER", LoggerDriver)

	HttpServerPort = envGetInt("HTTP_SERVER_PORT", HttpServerPort)
	HttpServerReadTimeout = envGetDuration("HTTP_SERVER_READ_TIMEOUT", HttpServerReadTimeout)
	HttpServerShutdownTimeout = envGetDuration("HTTP_SERVER_SHUTDOWN_TIMEOUT", HttpServerShutdownTimeout)
	HttpCorsAllowedOrigins = envGetStrings("HTTP_CORS_ALLOWED_ORIGINS", HttpCorsAllowedOrigins)

	PostgresDatabaseUrl = envGetString("POSTGRES_DATABASE_URL", PostgresDatabaseUrl)
	PostgresMaxConnections = envGetInt("POSTGRES_MAX_CONNECTIONS", PostgresMaxConnections)
	PostgresConnectTimeout = envGetDuration("POSTGRES_CONNECT_TIMEOUT", PostgresConnectTimeout)

	RedisAddresses = envGetStrings("REDIS_ADDRESSES", RedisAddresses)
	RedisUsername = envGetString("REDIS_USERNAME", RedisUsername)
	RedisPassword = envGetString("REDIS_PASSWORD", RedisPassword)
	RedisDatabase = envGetInt("REDIS_DATABASE", RedisDatabase)
	RedisConnectTimeout = envGetDuration("REDIS_CONNECT_TIMEOUT", RedisConnectTimeout)
	RedisCacheNamespace = envGetString("REDIS_CACHE_NAMESPACE", RedisCacheNamespace)
	RedisTtlIdentity = envGetDuration("REDIS_TTL_IDENTITY", RedisTtlIdentity)
	RedisTtlPagination = envGetDuration("REDIS_TTL_PAGINATION", RedisTtlPagination)
	RedisTtlRelation = envGetDuration("REDIS_TTL_RELATION", RedisTtlRelation)

	MinioEndpoint = envGetString("MINIO_ENDPOINT", envGetString("SEAWEEDFS_S3_ENDPOINT", MinioEndpoint))
	MinioAccessKey = envGetString("MINIO_ACCESS_KEY", envGetString("AWS_ACCESS_KEY_ID", MinioAccessKey))
	MinioSecretKey = envGetString("MINIO_SECRET_KEY", envGetString("AWS_SECRET_ACCESS_KEY", MinioSecretKey))
	MinioBucket = envGetString("MINIO_BUCKET", envGetString("SEAWEEDFS_BUCKET", MinioBucket))
	MinioUseSsl = envGetBool("MINIO_USE_SSL", MinioUseSsl)
	MinioRegion = envGetString("MINIO_REGION", MinioRegion)
	MinioConnectTimeout = envGetDuration("MINIO_CONNECT_TIMEOUT", MinioConnectTimeout)

	MqttBrokerUrl = envGetString("MQTT_BROKER_URL", MqttBrokerUrl)
	MqttClientId = envGetString("MQTT_CLIENT_ID", MqttClientId)
	MqttUsername = envGetString("MQTT_USERNAME", MqttUsername)
	MqttPassword = envGetString("MQTT_PASSWORD", MqttPassword)
	MqttConnectTimeout = envGetDuration("MQTT_CONNECT_TIMEOUT", MqttConnectTimeout)
	MqttKeepAlive = envGetDuration("MQTT_KEEP_ALIVE", MqttKeepAlive)
	MqttPingTimeout = envGetDuration("MQTT_PING_TIMEOUT", MqttPingTimeout)

	TokenAccessSecret = envGetString("TOKEN_ACCESS_SECRET", TokenAccessSecret)
	TokenRefreshSecret = envGetString("TOKEN_REFRESH_SECRET", TokenRefreshSecret)
	TokenAccessDuration = envGetDuration("TOKEN_ACCESS_DURATION", TokenAccessDuration)
	TokenRefreshDuration = envGetDuration("TOKEN_REFRESH_DURATION", TokenRefreshDuration)

	PasswordBcryptCost = envGetInt("PASSWORD_BCRYPT_COST", PasswordBcryptCost)
}

func envGetString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envGetStrings(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return fallback
	}

	return result
}

func envGetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	valInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return valInt
}

func envGetBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	valBool, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return valBool
}

func envGetDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func envGetLoggerFormat(key string, fallback domainmodels.LoggerFormat) domainmodels.LoggerFormat {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "plain":
		return domainmodels.LoggerFormatPlain
	case "json":
		return domainmodels.LoggerFormatJson
	default:
		return fallback
	}
}

func envGetLoggerLevel(key string, fallback domainmodels.LoggerLevel) domainmodels.LoggerLevel {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "error":
		return domainmodels.LoggerLevelError
	case "warn":
		return domainmodels.LoggerLevelWarn
	case "info":
		return domainmodels.LoggerLevelInfo
	case "debug":
		return domainmodels.LoggerLevelDebug
	default:
		return fallback
	}
}
