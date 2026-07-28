package infrastructurerepositoryactionlog

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var actionLogColumns = []string{
	"id",
	"execution_id",
	"action_id",
	"node_id",
	"action_status",
	"action_message",
	"payload",
	"executed_at",
	"created_at",
}

func (p *postgresImpl) queryCreate(
	executionId uuid.UUID,
	actionId uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus string,
	actionMessage *string,
	payload json.RawMessage,
	executedAt time.Time,
) (query string, args []any, err error) {
	return p.SqrD.Insert("action_logs").
		Columns("execution_id", "action_id", "node_id", "action_status", "action_message", "payload", "executed_at").
		Values(executionId, actionId, nodeId, actionStatus, actionMessage, payload, executedAt).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryUpdateStatusByExecutionId(
	executionId uuid.UUID,
	actionStatus string,
	actionMessage *string,
) (query string, args []any, err error) {
	return p.SqrD.Update("action_logs").
		Where(squirrel.Eq{"execution_id": executionId}).
		Set("action_status", actionStatus).
		Set("action_message", actionMessage).
		ToSql()
}

func (p *postgresImpl) queryReadByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(actionLogColumns...).
		From("action_logs")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("action_logs")

	baseQ, totalQ = applyActionLogFilters(baseQ, totalQ, executedAtStart, executedAtEnd, actionId, nodeId)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("executed_at DESC", "id ASC").
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("action_logs")

	if executedAtStart != nil {
		q = q.Where(squirrel.GtOrEq{"executed_at": *executedAtStart})
	}
	if executedAtEnd != nil {
		q = q.Where(squirrel.LtOrEq{"executed_at": *executedAtEnd})
	}
	if actionId != nil {
		q = q.Where(squirrel.Eq{"action_id": *actionId})
	}
	if nodeId != nil {
		q = q.Where(squirrel.Eq{"node_id": *nodeId})
	}

	return q.ToSql()
}

func applyActionLogFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if executedAtStart != nil {
		condition := squirrel.GtOrEq{"executed_at": *executedAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if executedAtEnd != nil {
		condition := squirrel.LtOrEq{"executed_at": *executedAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionId != nil {
		condition := squirrel.Eq{"action_id": *actionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeId != nil {
		condition := squirrel.Eq{"node_id": *nodeId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
