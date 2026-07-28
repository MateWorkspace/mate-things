package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PermissionResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type RoleResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type RolePermissionResponse struct {
	Id           string    `json:"id"`
	RoleId       string    `json:"role_id"`
	PermissionId string    `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    *string   `json:"created_by,omitempty"`
}

type RolePermissionDetailResponse struct {
	RolePermission RolePermissionResponse `json:"role_permission"`
	Role           RoleResponse           `json:"role"`
	Permission     PermissionResponse     `json:"permission"`
}

type UserResponse struct {
	Id          string          `json:"id"`
	RoleId      string          `json:"role_id"`
	Name        string          `json:"name"`
	Bio         string          `json:"bio"`
	Username    string          `json:"username"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type NodeClassResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type NodeResponse struct {
	Id          string          `json:"id"`
	NodeClassId string          `json:"node_class_id"`
	DeviceId    string          `json:"device_id"`
	DeviceInfo  string          `json:"device_info"`
	Name        string          `json:"name"`
	FirmwareId  string          `json:"firmware_id"`
	Description string          `json:"description"`
	IsConnected bool            `json:"is_connected"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type FirmwareResponse struct {
	Id          string          `json:"id"`
	NodeClassId string          `json:"node_class_id"`
	Name        string          `json:"name"`
	Size        int32           `json:"size"`
	Checksum    string          `json:"checksum"`
	BinaryPath  string          `json:"binary_path"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type FirmwareBinaryStatResponse struct {
	BinaryPath string `json:"binary_path"`
	Size       int32  `json:"size"`
	Checksum   string `json:"checksum"`
}

type FirmwareCreateResponse struct {
	Id         string `json:"id"`
	Size       int32  `json:"size"`
	Checksum   string `json:"checksum"`
	BinaryPath string `json:"binary_path"`
}

type PayloadSchemaResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Version     int32           `json:"version"`
	Definition  json.RawMessage `json:"definition"`
	ValidFrom   time.Time       `json:"valid_from"`
	ValidTo     *time.Time      `json:"valid_to,omitempty"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type ActionResponse struct {
	Id                   string          `json:"id"`
	NodeClassId          string          `json:"node_class_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	PayloadSchemaName    string          `json:"payload_schema_name"`
	PayloadSchemaVersion int32           `json:"payload_schema_version"`
	Preferences          json.RawMessage `json:"preferences"`
	AuditResponse
}

type ActionLogResponse struct {
	Id            int64                     `json:"id"`
	ExecutionId   string                    `json:"execution_id"`
	ActionId      string                    `json:"action_id"`
	NodeId        *string                   `json:"node_id,omitempty"`
	ActionStatus  domainmodels.ActionStatus `json:"action_status"`
	ActionMessage *string                   `json:"action_message,omitempty"`
	Payload       json.RawMessage           `json:"payload"`
	ExecutedAt    time.Time                 `json:"executed_at"`
	CreatedAt     time.Time                 `json:"created_at"`
}

type TelemetryRecordResponse struct {
	Id                   int64           `json:"id"`
	NodeDeviceId         string          `json:"node_device_id"`
	MetricName           string          `json:"metric_name"`
	PayloadSchemaName    string          `json:"payload_schema_name"`
	PayloadSchemaVersion int32           `json:"payload_schema_version"`
	Payload              json.RawMessage `json:"payload"`
	RecordedAt           time.Time       `json:"recorded_at"`
	CreatedAt            time.Time       `json:"created_at"`
}

type LoginResponse struct {
	User         UserResponse         `json:"user"`
	Role         RoleResponse         `json:"role"`
	Permissions  []PermissionResponse `json:"permissions"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token,omitempty"`
}

func Permission(permission domainmodels.Permission) PermissionResponse {
	return PermissionResponse{
		Id:          UUIDString(permission.Id),
		Name:        permission.Name,
		Description: permission.Description,
		Preferences: NormalizeJSON(permission.Preferences),
		AuditResponse: Audit(
			permission.CreatedAt,
			permission.UpdatedAt,
			permission.DeletedAt,
			permission.CreatedBy,
			permission.UpdatedBy,
			permission.DeletedBy,
		),
	}
}

func Permissions(permissions []domainmodels.Permission) []PermissionResponse {
	result := make([]PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, Permission(permission))
	}
	return result
}

func Role(role domainmodels.Role) RoleResponse {
	return RoleResponse{
		Id:          UUIDString(role.Id),
		Name:        role.Name,
		Description: role.Description,
		IsDefault:   role.IsDefault,
		Preferences: NormalizeJSON(role.Preferences),
		AuditResponse: Audit(
			role.CreatedAt,
			role.UpdatedAt,
			role.DeletedAt,
			role.CreatedBy,
			role.UpdatedBy,
			role.DeletedBy,
		),
	}
}

func Roles(roles []domainmodels.Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, Role(role))
	}
	return result
}

func RolePermission(rolePermission domainmodels.RolePermission) RolePermissionResponse {
	return RolePermissionResponse{
		Id:           UUIDString(rolePermission.Id),
		RoleId:       UUIDString(rolePermission.RoleId),
		PermissionId: UUIDString(rolePermission.PermissionId),
		CreatedAt:    rolePermission.CreatedAt,
		CreatedBy:    UUIDPtrString(rolePermission.CreatedBy),
	}
}

func RolePermissionDetail(rolePermission domainmodels.RolePermission, role domainmodels.Role, permission domainmodels.Permission) RolePermissionDetailResponse {
	return RolePermissionDetailResponse{
		RolePermission: RolePermission(rolePermission),
		Role:           Role(role),
		Permission:     Permission(permission),
	}
}

func RolePermissionDetails(
	rolePermissions []domainmodels.RolePermission,
	roles []domainmodels.Role,
	permissions []domainmodels.Permission,
) []RolePermissionDetailResponse {
	result := make([]RolePermissionDetailResponse, 0, len(rolePermissions))
	for i, rolePermission := range rolePermissions {
		var role domainmodels.Role
		if i < len(roles) {
			role = roles[i]
		}

		var permission domainmodels.Permission
		if i < len(permissions) {
			permission = permissions[i]
		}

		result = append(result, RolePermissionDetail(rolePermission, role, permission))
	}
	return result
}

func User(user domainmodels.User) UserResponse {
	return UserResponse{
		Id:          UUIDString(user.Id),
		RoleId:      UUIDString(user.RoleId),
		Name:        user.Name,
		Bio:         user.Bio,
		Username:    user.Username,
		Preferences: NormalizeJSON(user.Preferences),
		AuditResponse: Audit(
			user.CreatedAt,
			user.UpdatedAt,
			user.DeletedAt,
			user.CreatedBy,
			user.UpdatedBy,
			user.DeletedBy,
		),
	}
}

func Users(users []domainmodels.User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, User(user))
	}
	return result
}

func NodeClass(nodeClass domainmodels.NodeClass) NodeClassResponse {
	return NodeClassResponse{
		Id:          UUIDString(nodeClass.Id),
		Name:        nodeClass.Name,
		Description: nodeClass.Description,
		Preferences: NormalizeJSON(nodeClass.Preferences),
		AuditResponse: Audit(
			nodeClass.CreatedAt,
			nodeClass.UpdatedAt,
			nodeClass.DeletedAt,
			nodeClass.CreatedBy,
			nodeClass.UpdatedBy,
			nodeClass.DeletedBy,
		),
	}
}

func NodeClasses(nodeClasses []domainmodels.NodeClass) []NodeClassResponse {
	result := make([]NodeClassResponse, 0, len(nodeClasses))
	for _, nodeClass := range nodeClasses {
		result = append(result, NodeClass(nodeClass))
	}
	return result
}

func Node(node domainmodels.Node) NodeResponse {
	return NodeResponse{
		Id:          UUIDString(node.Id),
		NodeClassId: UUIDString(node.NodeClassId),
		DeviceId:    node.DeviceId,
		DeviceInfo:  node.DeviceInfo,
		Name:        node.Name,
		FirmwareId:  UUIDString(node.FirmwareId),
		Description: node.Description,
		IsConnected: node.IsConnected,
		Preferences: NormalizeJSON(node.Preferences),
		AuditResponse: Audit(
			node.CreatedAt,
			node.UpdatedAt,
			node.DeletedAt,
			node.CreatedBy,
			node.UpdatedBy,
			node.DeletedBy,
		),
	}
}

func Nodes(nodes []domainmodels.Node) []NodeResponse {
	result := make([]NodeResponse, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, Node(node))
	}
	return result
}

func Firmware(firmware domainmodels.Firmware) FirmwareResponse {
	return FirmwareResponse{
		Id:          UUIDString(firmware.Id),
		NodeClassId: UUIDString(firmware.NodeClassId),
		Name:        firmware.Name,
		Size:        firmware.Size,
		Checksum:    firmware.Checksum,
		BinaryPath:  firmware.BinaryPath,
		Preferences: NormalizeJSON(firmware.Preferences),
		AuditResponse: Audit(
			firmware.CreatedAt,
			firmware.UpdatedAt,
			firmware.DeletedAt,
			firmware.CreatedBy,
			firmware.UpdatedBy,
			firmware.DeletedBy,
		),
	}
}

func Firmwares(firmwares []domainmodels.Firmware) []FirmwareResponse {
	result := make([]FirmwareResponse, 0, len(firmwares))
	for _, firmware := range firmwares {
		result = append(result, Firmware(firmware))
	}
	return result
}

func PayloadSchema(payloadSchema domainmodels.PayloadSchema) PayloadSchemaResponse {
	return PayloadSchemaResponse{
		Id:          UUIDString(payloadSchema.Id),
		Name:        payloadSchema.Name,
		Version:     payloadSchema.Version,
		Definition:  NormalizeJSON(payloadSchema.Definition),
		ValidFrom:   payloadSchema.ValidFrom,
		ValidTo:     payloadSchema.ValidTo,
		Preferences: NormalizeJSON(payloadSchema.Preferences),
		AuditResponse: Audit(
			payloadSchema.CreatedAt,
			payloadSchema.UpdatedAt,
			payloadSchema.DeletedAt,
			payloadSchema.CreatedBy,
			payloadSchema.UpdatedBy,
			payloadSchema.DeletedBy,
		),
	}
}

func PayloadSchemas(payloadSchemas []domainmodels.PayloadSchema) []PayloadSchemaResponse {
	result := make([]PayloadSchemaResponse, 0, len(payloadSchemas))
	for _, payloadSchema := range payloadSchemas {
		result = append(result, PayloadSchema(payloadSchema))
	}
	return result
}

func Action(action domainmodels.Action) ActionResponse {
	return ActionResponse{
		Id:                   UUIDString(action.Id),
		NodeClassId:          UUIDString(action.NodeClassId),
		Name:                 action.Name,
		Description:          action.Description,
		PayloadSchemaName:    action.PayloadSchemaName,
		PayloadSchemaVersion: action.PayloadSchemaVersion,
		Preferences:          NormalizeJSON(action.Preferences),
		AuditResponse: Audit(
			action.CreatedAt,
			action.UpdatedAt,
			action.DeletedAt,
			action.CreatedBy,
			action.UpdatedBy,
			action.DeletedBy,
		),
	}
}

func Actions(actions []domainmodels.Action) []ActionResponse {
	result := make([]ActionResponse, 0, len(actions))
	for _, action := range actions {
		result = append(result, Action(action))
	}
	return result
}

func ActionLog(actionLog domainmodels.ActionLog) ActionLogResponse {
	return ActionLogResponse{
		Id:            actionLog.Id,
		ExecutionId:   UUIDString(actionLog.ExecutionId),
		ActionId:      UUIDString(actionLog.ActionId),
		NodeId:        UUIDPtrString(actionLog.NodeId),
		ActionStatus:  actionLog.ActionStatus,
		ActionMessage: actionLog.ActionMessage,
		Payload:       NormalizeJSON(actionLog.Payload),
		ExecutedAt:    actionLog.ExecutedAt,
		CreatedAt:     actionLog.CreatedAt,
	}
}

func ActionLogs(actionLogs []domainmodels.ActionLog) []ActionLogResponse {
	result := make([]ActionLogResponse, 0, len(actionLogs))
	for _, actionLog := range actionLogs {
		result = append(result, ActionLog(actionLog))
	}
	return result
}

func TelemetryRecord(telemetryRecord domainmodels.TelemetryRecord) TelemetryRecordResponse {
	return TelemetryRecordResponse{
		Id:                   telemetryRecord.Id,
		NodeDeviceId:         telemetryRecord.NodeDeviceId,
		MetricName:           telemetryRecord.MetricName,
		PayloadSchemaName:    telemetryRecord.PayloadSchemaName,
		PayloadSchemaVersion: telemetryRecord.PayloadSchemaVersion,
		Payload:              NormalizeJSON(telemetryRecord.Payload),
		RecordedAt:           telemetryRecord.RecordedAt,
		CreatedAt:            telemetryRecord.CreatedAt,
	}
}

func TelemetryRecords(telemetryRecords []domainmodels.TelemetryRecord) []TelemetryRecordResponse {
	result := make([]TelemetryRecordResponse, 0, len(telemetryRecords))
	for _, telemetryRecord := range telemetryRecords {
		result = append(result, TelemetryRecord(telemetryRecord))
	}
	return result
}
