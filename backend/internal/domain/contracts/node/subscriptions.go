package domaincontractsnode

import "context"

type Subscriptions interface {
	Registration(ctx context.Context) (err error)
	Status(ctx context.Context, nodeDeviceId string) (err error)
	Log(ctx context.Context, nodeDeviceId string) (err error)
	ActionAck(ctx context.Context, nodeDeviceId string) (err error)
	Telemetry(ctx context.Context, nodeDeviceId string) (err error)
	IrCapture(ctx context.Context, nodeDeviceId string) (err error)
	IrTransmitAck(ctx context.Context, nodeDeviceId string) (err error)
}
