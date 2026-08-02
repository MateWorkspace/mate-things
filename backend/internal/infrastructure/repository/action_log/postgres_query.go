package infrastructurerepositoryactionlog

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
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
	actionStatus *domainmodels.ActionStatus,
	page int,
	limit int,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	qualifiedColumns := make([]string, len(actionLogColumns))
	for i, col := range actionLogColumns {
		qualifiedColumns[i] = "action_logs." + col
	}

	// LEFT JOIN (not INNER) on both: node_id is nullable, and defensively,
	// a row referencing a since-deleted action/node should still surface
	// rather than silently vanish from the list. The count query stays
	// join-free - it never needs the joined name columns, only the same
	// WHERE filters, none of which reference an ambiguous column name.
	baseQ := p.SqrD.Select(qualifiedColumns...).
		Column("actions.name AS action_name").
		Column("nodes.device_id AS node_device_id").
		Column("nodes.name AS node_name").
		From("action_logs").
		LeftJoin("actions ON actions.id = action_logs.action_id").
		LeftJoin("nodes ON nodes.id = action_logs.node_id")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("action_logs")

	baseQ, totalQ = applyActionLogFilters(baseQ, totalQ, executedAtStart, executedAtEnd, actionId, nodeId, actionStatus)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("action_logs.executed_at DESC", "action_logs.id ASC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
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
	if actionStatus != nil {
		q = q.Where(squirrel.Eq{"action_status": string(*actionStatus)})
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
	actionStatus *domainmodels.ActionStatus,
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
	if actionStatus != nil {
		condition := squirrel.Eq{"action_status": string(*actionStatus)}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
