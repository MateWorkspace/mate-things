package infrastructurerepositoryrolepermission

import (
	"context"
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
) domaincontractsrepository.RolePermission {
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
	roleId uuid.UUID,
	permissionId uuid.UUID,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(roleId, permissionId, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create role permission query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create role permission", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "role_id_permission_id", Type: domainmodels.ErrTypeRolePermissionExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (rolePermission *domainmodels.RolePermission, role *domainmodels.Role, permission *domainmodels.Permission, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read role permission query", err)
	}

	rp, r, perm, err := scanPgxRolePermissionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("role permission not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read role permission", err)
	}

	return &rp, &r, &perm, nil
}

func (p *postgresImpl) ReadByRoleIdAndPermissionId(
	ctx context.Context,
	roleId uuid.UUID,
	permissionId uuid.UUID,
) (rolePermission *domainmodels.RolePermission, role *domainmodels.Role, permission *domainmodels.Permission, err error) {
	query, args, err := p.queryReadByRoleIdAndPermissionId(roleId, permissionId)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read role permission query", err)
	}

	rp, r, perm, err := scanPgxRolePermissionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("role permission not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read role permission", err)
	}

	return &rp, &r, &perm, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) (rolePermissions []domainmodels.RolePermission, roles []domainmodels.Role, permissions []domainmodels.Permission, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, roleId, permissionId)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read role permissions query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count role permissions", err)
	}
	if total == 0 {
		return []domainmodels.RolePermission{}, []domainmodels.Role{}, []domainmodels.Permission{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read role permissions", err)
	}
	defer rows.Close()

	rolePermissions = make([]domainmodels.RolePermission, 0, total)
	roles = make([]domainmodels.Role, 0, total)
	permissions = make([]domainmodels.Permission, 0, total)
	for rows.Next() {
		rp, r, perm, err := scanPgxRolePermissionJoined(rows)
		if err != nil {
			return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan role permission", err)
		}

		rolePermissions = append(rolePermissions, rp)
		roles = append(roles, r)
		permissions = append(permissions, perm)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to iterate role permissions", err)
	}

	return rolePermissions, roles, permissions, total, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID) (err error) {
	query, args, err := p.queryDeleteById(id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete role permission query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete role permission", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("role permission not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteByRoleIdAndPermissionId(
	ctx context.Context,
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteByRoleIdAndPermissionId(roleId, permissionId)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete role permission query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete role permission", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("role permission not found", nil)
	}

	return nil
}

func scanPgxRolePermissionJoined(row pgx.Row) (
	rolePermission domainmodels.RolePermission,
	role domainmodels.Role,
	permission domainmodels.Permission,
	err error,
) {
	err = row.Scan(
		&rolePermission.Id,
		&rolePermission.RoleId,
		&rolePermission.PermissionId,
		&rolePermission.CreatedAt,
		&rolePermission.CreatedBy,
		&role.Id,
		&role.Name,
		&role.Description,
		&role.IsDefault,
		&role.Preferences,
		&role.CreatedAt,
		&role.UpdatedAt,
		&role.DeletedAt,
		&role.CreatedBy,
		&role.UpdatedBy,
		&role.DeletedBy,
		&permission.Id,
		&permission.Name,
		&permission.Description,
		&permission.Preferences,
		&permission.CreatedAt,
		&permission.UpdatedAt,
		&permission.DeletedAt,
		&permission.CreatedBy,
		&permission.UpdatedBy,
		&permission.DeletedBy,
	)
	return
}
