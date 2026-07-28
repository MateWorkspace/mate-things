package domaincontractsnode

import "context"

type Subscriptions interface {
	Registration(ctx context.Context) (err error)
	Status(ctx context.Context, nodeDeviceId string) (err error)
	Log(ctx context.Context, nodeDeviceId string) (err error)
	ActionAck(ctx context.Context, nodeDeviceId string) (err error)
}
