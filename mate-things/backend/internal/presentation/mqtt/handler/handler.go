package presentationmqtthandler

import (
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

type Handler struct {
	Logger            domaincontractslogger.Leveled
	MessagingCallback domainusecasesnode.MessagingCallback
}

var H *Handler

func New(
	logger domaincontractslogger.Leveled,
	messagingCallback domainusecasesnode.MessagingCallback,
) *Handler {
	H = &Handler{
		Logger:            logger,
		MessagingCallback: messagingCallback,
	}
	return H
}
