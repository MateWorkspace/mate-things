package infrastructurerepositorytelemetryrecord

import (
	"context"
	"encoding/json"
	"time"

	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/ABA-Developer/nusapala-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.TelemetryRecord {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(
	ctx context.Context,
	nodeDeviceId string,
	metricName string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	payload json.RawMessage,
	recordedAt time.Time,
) (id int64, err error) {
	query, args, err := p.queryCreate(nodeDeviceId, metricName, payloadSchemaName, payloadSchemaVersion, payload, recordedAt)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build create telemetry record query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to create telemetry record", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	recordedAtStart *time.Time,
	recordedAtEnd *time.Time,
	nodeDeviceId *string,
	metricName *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (telemetryRecords []domainmodels.TelemetryRecord, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(recordedAtStart, recordedAtEnd, nodeDeviceId, metricName, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read telemetry records query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count telemetry records", err)
	}
	if total == 0 {
		return []domainmodels.TelemetryRecord{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read telemetry records", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxTelemetryRecords(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan telemetry records", err)
	}

	return items, total, nil
}

func (p *postgresImpl) DeleteByFilter(
	ctx context.Context,
	recordedAtStart *time.Time,
	recordedAtEnd *time.Time,
	nodeDeviceId *string,
	metricName *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (total int, err error) {
	query, args, err := p.queryDeleteByFilter(recordedAtStart, recordedAtEnd, nodeDeviceId, metricName, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build delete telemetry records query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to delete telemetry records", err)
	}

	return int(commandTag.RowsAffected()), nil
}
