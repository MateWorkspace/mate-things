package infrastructurerepositorytelemetryrecord

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
)

var telemetryRecordColumns = []string{
	"id",
	"node_device_id",
	"metric_name",
	"payload_schema_name",
	"payload_schema_version",
	"payload",
	"recorded_at",
	"created_at",
}

func (p *postgresImpl) queryCreate(
	nodeDeviceId string,
	metricName string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	payload json.RawMessage,
	recordedAt time.Time,
) (query string, args []any, err error) {
	return p.SqrD.Insert("telemetry_records").
		Columns("node_device_id", "metric_name", "payload_schema_name", "payload_schema_version", "payload", "recorded_at").
		Values(nodeDeviceId, metricName, payloadSchemaName, payloadSchemaVersion, payload, recordedAt).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadByFilter(
	recordedAtStart *time.Time,
	recordedAtEnd *time.Time,
	nodeDeviceId *string,
	metricName *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(telemetryRecordColumns...).
		From("telemetry_records")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("telemetry_records")

	baseQ, totalQ = applyTelemetryFilters(baseQ, totalQ, recordedAtStart, recordedAtEnd, nodeDeviceId, metricName, payloadSchemaName, payloadSchemaVersion)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("recorded_at DESC", "id ASC").
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteByFilter(
	recordedAtStart *time.Time,
	recordedAtEnd *time.Time,
	nodeDeviceId *string,
	metricName *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("telemetry_records")

	if recordedAtStart != nil {
		q = q.Where(squirrel.GtOrEq{"recorded_at": *recordedAtStart})
	}
	if recordedAtEnd != nil {
		q = q.Where(squirrel.LtOrEq{"recorded_at": *recordedAtEnd})
	}
	if nodeDeviceId != nil {
		q = q.Where(squirrel.Eq{"node_device_id": *nodeDeviceId})
	}
	if metricName != nil {
		q = q.Where(squirrel.Eq{"metric_name": *metricName})
	}
	if payloadSchemaName != nil {
		q = q.Where(squirrel.Eq{"payload_schema_name": *payloadSchemaName})
	}
	if payloadSchemaVersion != nil {
		q = q.Where(squirrel.Eq{"payload_schema_version": *payloadSchemaVersion})
	}

	return q.ToSql()
}

func applyTelemetryFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	recordedAtStart *time.Time,
	recordedAtEnd *time.Time,
	nodeDeviceId *string,
	metricName *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if recordedAtStart != nil {
		condition := squirrel.GtOrEq{"recorded_at": *recordedAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if recordedAtEnd != nil {
		condition := squirrel.LtOrEq{"recorded_at": *recordedAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeDeviceId != nil {
		condition := squirrel.Eq{"node_device_id": *nodeDeviceId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if metricName != nil {
		condition := squirrel.Eq{"metric_name": *metricName}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaName != nil {
		condition := squirrel.Eq{"payload_schema_name": *payloadSchemaName}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaVersion != nil {
		condition := squirrel.Eq{"payload_schema_version": *payloadSchemaVersion}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
