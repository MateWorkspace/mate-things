package infrastructurerepositorynodelog

import (
	"time"

	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

var nodeLogColumns = []string{
	"id",
	"node_device_id",
	"level",
	"tag",
	"message",
	"logged_at",
	"created_at",
}

func (p *postgresImpl) queryCreate(
	nodeDeviceId string,
	level domainmodels.NodeLogLevel,
	tag string,
	message string,
	loggedAt time.Time,
) (query string, args []any, err error) {
	return p.SqrD.Insert("node_logs").
		Columns("node_device_id", "level", "tag", "message", "logged_at").
		Values(nodeDeviceId, string(level), tag, message, loggedAt).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadByFilter(
	loggedAtStart *time.Time,
	loggedAtEnd *time.Time,
	nodeDeviceId *string,
	level *domainmodels.NodeLogLevel,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(nodeLogColumns...).
		From("node_logs")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("node_logs")

	baseQ, totalQ = applyNodeLogFilters(baseQ, totalQ, loggedAtStart, loggedAtEnd, nodeDeviceId, level)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("logged_at DESC", "id ASC").
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteByFilter(
	loggedAtStart *time.Time,
	loggedAtEnd *time.Time,
	nodeDeviceId *string,
	level *domainmodels.NodeLogLevel,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("node_logs")

	if loggedAtStart != nil {
		q = q.Where(squirrel.GtOrEq{"logged_at": *loggedAtStart})
	}
	if loggedAtEnd != nil {
		q = q.Where(squirrel.LtOrEq{"logged_at": *loggedAtEnd})
	}
	if nodeDeviceId != nil {
		q = q.Where(squirrel.Eq{"node_device_id": *nodeDeviceId})
	}
	if level != nil {
		q = q.Where(squirrel.Eq{"level": string(*level)})
	}

	return q.ToSql()
}

func applyNodeLogFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	loggedAtStart *time.Time,
	loggedAtEnd *time.Time,
	nodeDeviceId *string,
	level *domainmodels.NodeLogLevel,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if loggedAtStart != nil {
		condition := squirrel.GtOrEq{"logged_at": *loggedAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if loggedAtEnd != nil {
		condition := squirrel.LtOrEq{"logged_at": *loggedAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeDeviceId != nil {
		condition := squirrel.Eq{"node_device_id": *nodeDeviceId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if level != nil {
		condition := squirrel.Eq{"level": string(*level)}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
