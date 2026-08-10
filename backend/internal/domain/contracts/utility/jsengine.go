package domaincontractsutility

import "time"

type JSEngine interface {
	RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)
}
