package infrastructurerepositoryshared

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/jackc/pgx/v5"
)

func ScanPgxPermission(row pgx.Row) (domainmodels.Permission, error) {
	var item domainmodels.Permission
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxPermissions(rows pgx.Rows) ([]domainmodels.Permission, error) {
	items := make([]domainmodels.Permission, 0)
	for rows.Next() {
		item, err := ScanPgxPermission(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxRole(row pgx.Row) (domainmodels.Role, error) {
	var item domainmodels.Role
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.IsDefault,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxRoles(rows pgx.Rows) ([]domainmodels.Role, error) {
	items := make([]domainmodels.Role, 0)
	for rows.Next() {
		item, err := ScanPgxRole(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxRolePermission(row pgx.Row) (domainmodels.RolePermission, error) {
	var item domainmodels.RolePermission
	err := row.Scan(
		&item.Id,
		&item.RoleId,
		&item.PermissionId,
		&item.CreatedAt,
		&item.CreatedBy,
	)
	return item, err
}

func ScanPgxUser(row pgx.Row) (domainmodels.User, error) {
	var item domainmodels.User
	err := row.Scan(
		&item.Id,
		&item.RoleId,
		&item.Name,
		&item.Bio,
		&item.Username,
		&item.PasswordHash,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxUsers(rows pgx.Rows) ([]domainmodels.User, error) {
	items := make([]domainmodels.User, 0)
	for rows.Next() {
		item, err := ScanPgxUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxNodeClass(row pgx.Row) (domainmodels.NodeClass, error) {
	var item domainmodels.NodeClass
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxNodeClasses(rows pgx.Rows) ([]domainmodels.NodeClass, error) {
	items := make([]domainmodels.NodeClass, 0)
	for rows.Next() {
		item, err := ScanPgxNodeClass(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxFirmware(row pgx.Row) (domainmodels.Firmware, error) {
	var item domainmodels.Firmware
	err := row.Scan(
		&item.Id,
		&item.NodeClassId,
		&item.Name,
		&item.Size,
		&item.Checksum,
		&item.BinaryPath,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxFirmwares(rows pgx.Rows) ([]domainmodels.Firmware, error) {
	items := make([]domainmodels.Firmware, 0)
	for rows.Next() {
		item, err := ScanPgxFirmware(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxNode(row pgx.Row) (domainmodels.Node, error) {
	var item domainmodels.Node
	err := row.Scan(
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
	return item, err
}

func ScanPgxNodes(rows pgx.Rows) ([]domainmodels.Node, error) {
	items := make([]domainmodels.Node, 0)
	for rows.Next() {
		item, err := ScanPgxNode(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxPayloadSchema(row pgx.Row) (domainmodels.PayloadSchema, error) {
	var item domainmodels.PayloadSchema
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Version,
		&item.Definition,
		&item.ValidFrom,
		&item.ValidTo,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxPayloadSchemas(rows pgx.Rows) ([]domainmodels.PayloadSchema, error) {
	items := make([]domainmodels.PayloadSchema, 0)
	for rows.Next() {
		item, err := ScanPgxPayloadSchema(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxAction(row pgx.Row) (domainmodels.Action, error) {
	var item domainmodels.Action
	err := row.Scan(
		&item.Id,
		&item.NodeClassId,
		&item.Name,
		&item.Description,
		&item.PayloadSchemaName,
		&item.PayloadSchemaVersion,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxActions(rows pgx.Rows) ([]domainmodels.Action, error) {
	items := make([]domainmodels.Action, 0)
	for rows.Next() {
		item, err := ScanPgxAction(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxTelemetryRecord(row pgx.Row) (domainmodels.TelemetryRecord, error) {
	var item domainmodels.TelemetryRecord
	err := row.Scan(
		&item.Id,
		&item.NodeDeviceId,
		&item.MetricName,
		&item.PayloadSchemaName,
		&item.PayloadSchemaVersion,
		&item.Payload,
		&item.RecordedAt,
		&item.CreatedAt,
	)
	return item, err
}

func ScanPgxTelemetryRecords(rows pgx.Rows) ([]domainmodels.TelemetryRecord, error) {
	items := make([]domainmodels.TelemetryRecord, 0)
	for rows.Next() {
		item, err := ScanPgxTelemetryRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxActionLog(row pgx.Row) (domainmodels.ActionLog, error) {
	var item domainmodels.ActionLog
	err := row.Scan(
		&item.Id,
		&item.ExecutionId,
		&item.ActionId,
		&item.NodeId,
		&item.ActionStatus,
		&item.ActionMessage,
		&item.Payload,
		&item.ExecutedAt,
		&item.CreatedAt,
	)
	return item, err
}

func ScanPgxActionLogs(rows pgx.Rows) ([]domainmodels.ActionLog, error) {
	items := make([]domainmodels.ActionLog, 0)
	for rows.Next() {
		item, err := ScanPgxActionLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
