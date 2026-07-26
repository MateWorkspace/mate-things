package infrastructurecacheshared

import "time"

type TtlConfig struct {
	Identity   time.Duration
	Pagination time.Duration
	Relation   time.Duration
}

func DefaultTtlConfig() TtlConfig {
	return TtlConfig{
		Identity:   30 * time.Minute,
		Pagination: 5 * time.Minute,
		Relation:   10 * time.Minute,
	}
}
