package domaincontractsrepository

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeLog interface {
	Create(
		ctx context.Context,
		nodeDeviceId string,
		level domainmodels.NodeLogLevel,
		tag string,
		message string,
		loggedAt time.Time,
	) (id int64, err error)

	ReadByFilter(
		ctx context.Context,
		loggedAtStart *time.Time,
		loggedAtEnd *time.Time,
		nodeDeviceId *string,
		level *domainmodels.NodeLogLevel,
	) (nodeLogs []domainmodels.NodeLog, total int, err error)

	DeleteByFilter(
		ctx context.Context,
		loggedAtStart *time.Time,
		loggedAtEnd *time.Time,
		nodeDeviceId *string,
		level *domainmodels.NodeLogLevel,
	) (total int, err error)
}
