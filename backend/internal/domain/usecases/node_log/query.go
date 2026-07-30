package domainusecasesnodelog

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type Query interface {
	ReadByFilter(ctx context.Context, request ReadNodeLogByFilterRequest) ([]domainmodels.NodeLog, int, error)
	DeleteByFilter(ctx context.Context, request DeleteNodeLogByFilterRequest) (int, error)
}

type ReadNodeLogByFilterRequest struct {
	LoggedAtStart *time.Time
	LoggedAtEnd   *time.Time
	NodeDeviceId  *string
	Level         *domainmodels.NodeLogLevel
}

type DeleteNodeLogByFilterRequest struct {
	LoggedAtStart *time.Time
	LoggedAtEnd   *time.Time
	NodeDeviceId  *string
	Level         *domainmodels.NodeLogLevel
}
