package presentationmqtthandler

import (
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

type Handler struct {
	Logger                domaincontractslogger.Leveled
	MessagingCallback     domainusecasesnode.MessagingCallback
	InfraredRecordSession domainusecasesinfrared.RecordSessionManagement
}

var H *Handler

func New(
	logger domaincontractslogger.Leveled,
	messagingCallback domainusecasesnode.MessagingCallback,
	infraredRecordSession domainusecasesinfrared.RecordSessionManagement,
) *Handler {
	H = &Handler{
		Logger:                logger,
		MessagingCallback:     messagingCallback,
		InfraredRecordSession: infraredRecordSession,
	}
	return H
}
