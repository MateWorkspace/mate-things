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

	BaseUrl string = "http://127.0.0.1:8080"

	HttpServerReadTimeout     time.Duration = 30 * time.Second
	HttpServerShutdownTimeout time.Duration = 10 * time.Second
	HttpCorsAllowedOrigins    []string      = []string{"*"}

	PostgresHost           string        = "127.0.0.1"
	PostgresPort           int           = 5432
	PostgresUsername       string        = "postgres"
	PostgresPassword       string        = "postgres"
	PostgresDatabase       string        = "nusapala_things"
	PostgresSslMode        string        = "disable"
	PostgresMaxConnections int           = 20
	PostgresConnectTimeout time.Duration = 10 * time.Second

	RedisAddresses      []string      = []string{"127.0.0.1:6379"}
	RedisPassword       string        = ""
	RedisDatabase       int           = 0
	RedisConnectTimeout time.Duration = 10 * time.Second
	RedisCacheNamespace string        = "nusapala-things"
	RedisTtlIdentity    time.Duration = 30 * time.Minute
	RedisTtlPagination  time.Duration = 5 * time.Minute
	RedisTtlRelation    time.Duration = 10 * time.Minute

	MinioEndpoint        string        = "127.0.0.1:8333"
	MinioAccessKey       string        = "nusapala"
	MinioSecretKey       string        = "nusapala-secret"
	MinioBucket          string        = "nusapala-things"
	MinioUseSsl          bool          = false
	MinioRegion          string        = ""
	MinioConnectTimeout  time.Duration = 10 * time.Second
	MinioPresignDuration time.Duration = 5 * time.Minute

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

	LlmEncryptionKey string = ""
)

func LoadEnv() {
	LoggerFormat = envGetLoggerFormat("BE_LOGGER_FORMAT", LoggerFormat)
	LoggerLevel = envGetLoggerLevel("BE_LOGGER_LEVEL", LoggerLevel)
	LoggerDriver = envGetString("BE_LOGGER_DRIVER", LoggerDriver)

	BaseUrl = envGetString("BE_BASE_URL", BaseUrl)

	HttpServerReadTimeout = envGetDuration("BE_HTTP_SERVER_READ_TIMEOUT", HttpServerReadTimeout)
	HttpServerShutdownTimeout = envGetDuration("BE_HTTP_SERVER_SHUTDOWN_TIMEOUT", HttpServerShutdownTimeout)
	HttpCorsAllowedOrigins = envGetStrings("BE_HTTP_CORS_ALLOWED_ORIGINS", HttpCorsAllowedOrigins)

	PostgresHost = envGetString("BE_POSTGRES_HOST", PostgresHost)
	PostgresPort = envGetInt("BE_POSTGRES_PORT", PostgresPort)
	PostgresUsername = envGetString("BE_POSTGRES_USERNAME", PostgresUsername)
	PostgresPassword = envGetString("BE_POSTGRES_PASSWORD", PostgresPassword)
	PostgresDatabase = envGetString("BE_POSTGRES_DATABASE", PostgresDatabase)
	PostgresSslMode = envGetString("BE_POSTGRES_SSL_MODE", PostgresSslMode)
	PostgresMaxConnections = envGetInt("BE_POSTGRES_MAX_CONNECTIONS", PostgresMaxConnections)
	PostgresConnectTimeout = envGetDuration("BE_POSTGRES_CONNECT_TIMEOUT", PostgresConnectTimeout)

	RedisAddresses = envGetStrings("BE_REDIS_ADDRESSES", RedisAddresses)
	RedisPassword = envGetString("BE_REDIS_PASSWORD", RedisPassword)
	RedisDatabase = envGetInt("BE_REDIS_DATABASE", RedisDatabase)
	RedisConnectTimeout = envGetDuration("BE_REDIS_CONNECT_TIMEOUT", RedisConnectTimeout)
	RedisCacheNamespace = envGetString("BE_REDIS_CACHE_NAMESPACE", RedisCacheNamespace)
	RedisTtlIdentity = envGetDuration("BE_REDIS_TTL_IDENTITY", RedisTtlIdentity)
	RedisTtlPagination = envGetDuration("BE_REDIS_TTL_PAGINATION", RedisTtlPagination)
	RedisTtlRelation = envGetDuration("BE_REDIS_TTL_RELATION", RedisTtlRelation)

	MinioEndpoint = envGetString("BE_MINIO_ENDPOINT", envGetString("BE_SEAWEEDFS_S3_ENDPOINT", MinioEndpoint))
	MinioAccessKey = envGetString("BE_MINIO_ACCESS_KEY", envGetString("BE_AWS_ACCESS_KEY_ID", MinioAccessKey))
	MinioSecretKey = envGetString("BE_MINIO_SECRET_KEY", envGetString("BE_AWS_SECRET_ACCESS_KEY", MinioSecretKey))
	MinioBucket = envGetString("BE_MINIO_BUCKET", envGetString("BE_SEAWEEDFS_BUCKET", MinioBucket))
	MinioUseSsl = envGetBool("BE_MINIO_USE_SSL", MinioUseSsl)
	MinioRegion = envGetString("BE_MINIO_REGION", MinioRegion)
	MinioConnectTimeout = envGetDuration("BE_MINIO_CONNECT_TIMEOUT", MinioConnectTimeout)
	MinioPresignDuration = envGetDuration("BE_MINIO_PRESIGN_DURATION", MinioPresignDuration)

	MqttBrokerUrl = envGetString("BE_MQTT_BROKER_URL", MqttBrokerUrl)
	MqttClientId = envGetString("BE_MQTT_CLIENT_ID", MqttClientId)
	MqttUsername = envGetString("BE_MQTT_USERNAME", MqttUsername)
	MqttPassword = envGetString("BE_MQTT_PASSWORD", MqttPassword)
	MqttConnectTimeout = envGetDuration("BE_MQTT_CONNECT_TIMEOUT", MqttConnectTimeout)
	MqttKeepAlive = envGetDuration("BE_MQTT_KEEP_ALIVE", MqttKeepAlive)
	MqttPingTimeout = envGetDuration("BE_MQTT_PING_TIMEOUT", MqttPingTimeout)

	TokenAccessSecret = envGetString("BE_TOKEN_ACCESS_SECRET", TokenAccessSecret)
	TokenRefreshSecret = envGetString("BE_TOKEN_REFRESH_SECRET", TokenRefreshSecret)
	TokenAccessDuration = envGetDuration("BE_TOKEN_ACCESS_DURATION", TokenAccessDuration)
	TokenRefreshDuration = envGetDuration("BE_TOKEN_REFRESH_DURATION", TokenRefreshDuration)

	PasswordBcryptCost = envGetInt("BE_PASSWORD_BCRYPT_COST", PasswordBcryptCost)

	LlmEncryptionKey = envGetString("BE_LLM_ENCRYPTION_KEY", LlmEncryptionKey)
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
