package presentationhttpresponse

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeLogResponse struct {
	Id           int64     `json:"id" example:"918273"`
	NodeDeviceId string    `json:"node_device_id" example:"AC276E5E030C"`
	Level        string    `json:"level" example:"INFO"`
	Tag          string    `json:"tag" example:"wifi_manager"`
	Message      string    `json:"message" example:"WiFi connected, IP: 192.168.1.42"`
	LoggedAt     time.Time `json:"logged_at" example:"2026-07-30T14:03:12.045Z"`
	CreatedAt    time.Time `json:"created_at" example:"2026-07-30T14:03:12.100Z"`
}

func NodeLog(nodeLog domainmodels.NodeLog) NodeLogResponse {
	return NodeLogResponse{
		Id:           nodeLog.Id,
		NodeDeviceId: nodeLog.NodeDeviceId,
		Level:        string(nodeLog.Level),
		Tag:          nodeLog.Tag,
		Message:      nodeLog.Message,
		LoggedAt:     nodeLog.LoggedAt,
		CreatedAt:    nodeLog.CreatedAt,
	}
}

func NodeLogs(nodeLogs []domainmodels.NodeLog) []NodeLogResponse {
	result := make([]NodeLogResponse, 0, len(nodeLogs))
	for _, nodeLog := range nodeLogs {
		result = append(result, NodeLog(nodeLog))
	}
	return result
}
