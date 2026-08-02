package applicationactionhistory

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
)

type usecase struct {
	actionLog domaincontractsrepository.ActionLog
	logger    domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	actionLog domaincontractsrepository.ActionLog,
	logger domaincontractslogger.Leveled,
) domainusecasesaction.History {
	return &usecase{
		actionLog: actionLog,
		logger:    logger,
	}
}

func (u *usecase) ReadByFilter(
	ctx context.Context,
	request domainusecasesaction.ReadActionLogsByFilterRequest,
) ([]domainmodels.ActionLogListItem, int, error) {
	const tag = "action/history/ReadByFilter"

	actionLogs, total, err := u.actionLog.ReadByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
		request.ActionStatus,
		request.Page,
		request.Limit,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action logs", domainmodels.LoggerMeta{
			"err":       err,
			"action_id": request.ActionId,
			"node_id":   request.NodeId,
		})
		return nil, 0, err
	}

	return actionLogs, total, nil
}

func (u *usecase) DeleteByFilter(
	ctx context.Context,
	request domainusecasesaction.DeleteActionLogsByFilterRequest,
) (int, error) {
	const tag = "action/history/DeleteByFilter"

	total, err := u.actionLog.DeleteByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
		request.ActionStatus,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to delete action logs", domainmodels.LoggerMeta{
			"err":       err,
			"action_id": request.ActionId,
			"node_id":   request.NodeId,
		})
		return 0, err
	}

	return total, nil
}
