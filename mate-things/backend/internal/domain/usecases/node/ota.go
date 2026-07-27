package domainusecasesnode

import (
	"context"

	"github.com/google/uuid"
)

type Ota interface {
	DispatchByNodeId(ctx context.Context, request DispatchOtaByNodeIdRequest) error
	DispatchByNodeDeviceId(ctx context.Context, request DispatchOtaByNodeDeviceIdRequest) error
}

type DispatchOtaByNodeIdRequest struct {
	NodeId      uuid.UUID
	FirmwareId  uuid.UUID
	FirmwareUrl string
	ActorId     *uuid.UUID
}

type DispatchOtaByNodeDeviceIdRequest struct {
	NodeDeviceId string
	FirmwareId   uuid.UUID
	FirmwareUrl  string
	ActorId      *uuid.UUID
}
