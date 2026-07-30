package applicationnodelogquery

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnodelog "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node_log"
)

type usecase struct {
	nodeLog domaincontractsrepository.NodeLog
	logger  domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	nodeLog domaincontractsrepository.NodeLog,
	logger domaincontractslogger.Leveled,
) domainusecasesnodelog.Query {
	return &usecase{
		nodeLog: nodeLog,
		logger:  logger,
	}
}

func (u *usecase) ReadByFilter(
	ctx context.Context,
	request domainusecasesnodelog.ReadNodeLogByFilterRequest,
) ([]domainmodels.NodeLog, int, error) {
	const tag = "node_log/query/ReadByFilter"

	nodeLogs, total, err := u.nodeLog.ReadByFilter(
		ctx,
		request.LoggedAtStart,
		request.LoggedAtEnd,
		request.NodeDeviceId,
		request.Level,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node logs", domainmodels.LoggerMeta{
			"err":            err,
			"node_device_id": request.NodeDeviceId,
			"level":          request.Level,
		})
		return nil, 0, err
	}

	return nodeLogs, total, nil
}

func (u *usecase) DeleteByFilter(
	ctx context.Context,
	request domainusecasesnodelog.DeleteNodeLogByFilterRequest,
) (int, error) {
	const tag = "node_log/query/DeleteByFilter"

	total, err := u.nodeLog.DeleteByFilter(
		ctx,
		request.LoggedAtStart,
		request.LoggedAtEnd,
		request.NodeDeviceId,
		request.Level,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to delete node logs", domainmodels.LoggerMeta{
			"err":            err,
			"node_device_id": request.NodeDeviceId,
			"level":          request.Level,
		})
		return 0, err
	}

	return total, nil
}
