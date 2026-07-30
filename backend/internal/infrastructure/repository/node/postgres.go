package infrastructurerepositorynode

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.Node {
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
	nodeClassId uuid.UUID,
	deviceId string,
	deviceInfo string,
	name string,
	firmwareId uuid.UUID,
	description *string,
	isConnected bool,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(nodeClassId, deviceId, deviceInfo, name, firmwareId, description, isConnected, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create node query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "device_id", Type: domainmodels.ErrTypeNodeDeviceIdExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) UpsertRegistration(
	ctx context.Context,
	deviceId string,
	deviceInfo string,
	firmwareName string,
) (node *domainmodels.Node, created bool, err error) {
	query, args, err := p.queryUpsertRegistration(deviceId, deviceInfo, firmwareName)
	if err != nil {
		return nil, false, infrastructurerepositoryshared.QueryBuildError("failed to build upsert node registration query", err)
	}

	var item domainmodels.Node
	row := p.Dt.QueryRow(ctx, query, args...)
	err = row.Scan(
		&created,
		&item.Id,
		&item.NodeClassId,
		&item.DeviceId,
		&item.DeviceInfo,
		&item.Name,
		&item.FirmwareId,
		&item.Description,
		&item.IsConnected,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, infrastructurerepositoryshared.NotFound("firmware not found", err)
		}
		return nil, false, infrastructurerepositoryshared.MapPgxError(
			"failed to upsert node registration", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "device_id", Type: domainmodels.ErrTypeNodeDeviceIdExists},
		)
	}

	return &item, created, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (node *domainmodels.Node, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxNode(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("node not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByDeviceId(ctx context.Context, deviceId string) (node *domainmodels.Node, err error) {
	query, args, err := p.queryReadByDeviceId(deviceId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxNode(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("node not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	firmwareId *uuid.UUID,
) (nodes []domainmodels.Node, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, nodeClassId, firmwareId)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read nodes query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count nodes", err)
	}
	if total == 0 {
		return []domainmodels.Node{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read nodes", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxNodes(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan nodes", err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	nodeClassId *uuid.UUID,
	deviceId *string,
	deviceInfo *string,
	name *string,
	firmwareId *uuid.UUID,
	description *string,
	isConnected *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, nodeClassId, deviceId, deviceInfo, name, firmwareId, description, isConnected, preferences, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update node query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update node", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "device_id", Type: domainmodels.ErrTypeNodeDeviceIdExists},
		)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node not found", nil)
	}

	return nil
}
