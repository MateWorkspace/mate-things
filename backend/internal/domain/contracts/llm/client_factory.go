package domaincontractsllm

import "context"

type ClientFactory interface {
	Current(ctx context.Context) (Client, error)
}
