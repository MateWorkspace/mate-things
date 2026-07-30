package domainmodels

import "time"

type NodeLogLevel string

const (
	NodeLogLevelNone  NodeLogLevel = "NONE"
	NodeLogLevelError NodeLogLevel = "ERROR"
	NodeLogLevelWarn  NodeLogLevel = "WARN"
	NodeLogLevelInfo  NodeLogLevel = "INFO"
	NodeLogLevelDebug NodeLogLevel = "DEBUG"
)

type NodeLog struct {
	Id           int64        `db:"id" json:"id"`
	NodeDeviceId string       `db:"node_device_id" json:"node_device_id"`
	Level        NodeLogLevel `db:"level" json:"level"`
	Tag          string       `db:"tag" json:"tag"`
	Message      string       `db:"message" json:"message"`
	LoggedAt     time.Time    `db:"logged_at" json:"logged_at"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}
