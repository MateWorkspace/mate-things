package applicationnodemessagingcallback

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	domainusecasestelemetry "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/telemetry"
)

const resubscribePageLimit = 100

type usecase struct {
	node                 domainusecasesrepocache.Node
	actionLog            domaincontractsrepository.ActionLog
	nodeLog              domaincontractsrepository.NodeLog
	firmwareConfigParams domaincontractsrepository.FirmwareConfigParameter
	nodeConfigValues     domaincontractsrepository.NodeConfigValue
	telemetryIngestion   domainusecasestelemetry.Ingestion
	publisher            domaincontractsnode.Publish
	subscriptions        domaincontractsnode.Subscriptions
	logger               domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	actionLog domaincontractsrepository.ActionLog,
	nodeLog domaincontractsrepository.NodeLog,
	firmwareConfigParams domaincontractsrepository.FirmwareConfigParameter,
	nodeConfigValues domaincontractsrepository.NodeConfigValue,
	telemetryIngestion domainusecasestelemetry.Ingestion,
	publisher domaincontractsnode.Publish,
	subscriptions domaincontractsnode.Subscriptions,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.MessagingCallback {
	return &usecase{
		node:                 node,
		actionLog:            actionLog,
		nodeLog:              nodeLog,
		firmwareConfigParams: firmwareConfigParams,
		nodeConfigValues:     nodeConfigValues,
		telemetryIngestion:   telemetryIngestion,
		publisher:            publisher,
		subscriptions:        subscriptions,
		logger:               logger,
	}
}

func (u *usecase) Register(ctx context.Context, request domainusecasesnode.RegisterNodeMessageRequest) error {
	const tag = "node/messaging_callback/Register"

	deviceId, err := applicationshared.RequiredDeviceId(request.DeviceId, "device_id")
	if err != nil {
		u.logger.Warn(ctx, tag, "invalid device_id in registration", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	node, created, err := u.node.UpsertRegistration(ctx, deviceId, request.DeviceInfo, request.FirmwareName)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to upsert node registration", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"firmware_name": request.FirmwareName,
		})
		return err
	}

	if err := u.syncReportedConfig(ctx, node, request.Config); err != nil {
		u.logger.Error(ctx, tag, "failed to sync reported node config", domainmodels.LoggerMeta{
			"err":         err,
			"device_id":   node.DeviceId,
			"node_id":     node.Id,
			"firmware_id": node.FirmwareId,
		})
		return err
	}

	if created {
		if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
			u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
				"err":       err,
				"device_id": node.DeviceId,
				"node_id":   node.Id,
			})
			return err
		}
	}

	if err := u.publisher.RegistrationAck(ctx, node.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to publish registration ack", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": node.DeviceId,
			"node_id":   node.Id,
		})
		return err
	}

	return nil
}

func (u *usecase) Status(ctx context.Context, request domainusecasesnode.NodeStatusMessageRequest) error {
	const tag = "node/messaging_callback/Status"

	node, err := u.node.ReadByDeviceId(ctx, request.DeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	if err := u.node.UpdateById(ctx, node.Id, nil, nil, nil, nil, nil, nil, &request.IsConnected, nil, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to update node status", domainmodels.LoggerMeta{
			"err":          err,
			"device_id":    request.DeviceId,
			"node_id":      node.Id,
			"is_connected": request.IsConnected,
		})
		return err
	}

	return nil
}

func (u *usecase) ActionAck(ctx context.Context, request domainusecasesnode.NodeActionAckMessageRequest) error {
	const tag = "node/messaging_callback/ActionAck"

	switch request.ActionStatus {
	case domainmodels.ActionStatusSuccess, domainmodels.ActionStatusFailed:
	default:
		err := domainmodels.NewError("invalid action ack status", domainmodels.ErrTypeValidation, nil)
		u.logger.Error(ctx, tag, "invalid action ack status", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"execution_id":  request.ExecutionId,
			"action_status": request.ActionStatus,
		})
		return err
	}

	if _, err := u.node.ReadByDeviceId(ctx, request.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":          err,
			"device_id":    request.DeviceId,
			"execution_id": request.ExecutionId,
		})
		return err
	}

	if err := u.actionLog.UpdateStatusByExecutionId(ctx, request.ExecutionId, request.ActionStatus, request.ActionMessage); err != nil {
		u.logger.Error(ctx, tag, "failed to update action log status", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"execution_id":  request.ExecutionId,
			"action_status": request.ActionStatus,
		})
		return err
	}

	return nil
}

func (u *usecase) Log(ctx context.Context, request domainusecasesnode.NodeLogMessageRequest) error {
	const tag = "node/messaging_callback/Log"

	level, logTag, message, loggedAt := parseLogLine(string(request.Payload))

	if _, err := u.nodeLog.Create(ctx, request.DeviceId, level, logTag, message, loggedAt); err != nil {
		u.logger.Error(ctx, tag, "failed to store node log", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	return nil
}

func (u *usecase) Telemetry(ctx context.Context, request domainusecasesnode.NodeTelemetryMessageRequest) error {
	const tag = "node/messaging_callback/Telemetry"

	if _, err := u.telemetryIngestion.Record(ctx, domainusecasestelemetry.RecordTelemetryRequest{
		NodeDeviceId:         request.DeviceId,
		MetricName:           request.MetricName,
		PayloadSchemaName:    request.PayloadSchemaName,
		PayloadSchemaVersion: request.PayloadSchemaVersion,
		Payload:              request.Payload,
		RecordedAt:           request.RecordedAt,
	}); err != nil {
		u.logger.Error(ctx, tag, "failed to record telemetry", domainmodels.LoggerMeta{
			"err":                    err,
			"device_id":              request.DeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
		})
		return err
	}

	return nil
}

func (u *usecase) Resubscribe(ctx context.Context) error {
	const tag = "node/messaging_callback/Resubscribe"

	if err := u.subscriptions.Registration(ctx); err != nil {
		u.logger.Error(ctx, tag, "failed to subscribe registration topic", domainmodels.LoggerMeta{
			"err": err,
		})
		return err
	}

	page := 1
	for {
		nodes, total, err := u.node.ReadByPagination(ctx, page, resubscribePageLimit, nil, nil, nil)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read nodes", domainmodels.LoggerMeta{
				"err":   err,
				"page":  page,
				"limit": resubscribePageLimit,
			})
			return err
		}
		for _, node := range nodes {
			if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
				u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
					"err":       err,
					"device_id": node.DeviceId,
					"node_id":   node.Id,
				})
				return err
			}
		}

		if page*resubscribePageLimit >= total {
			break
		}
		page++
	}

	return nil
}

func (u *usecase) subscribeNode(ctx context.Context, deviceId string) error {
	if err := u.subscriptions.ActionAck(ctx, deviceId); err != nil {
		return err
	}
	if err := u.subscriptions.Status(ctx, deviceId); err != nil {
		return err
	}
	if err := u.subscriptions.Log(ctx, deviceId); err != nil {
		return err
	}
	if err := u.subscriptions.Telemetry(ctx, deviceId); err != nil {
		return err
	}
	return nil
}

// syncReportedConfig upserts node_config_values from what the device itself
// reported at registration - the device is the source of truth for its own
// live config, so this runs on every registration (not just first-time
// creation), keeping the DB from drifting away from a value set out-of-band
// (e.g. via BLE) since the node's last reconnect. A key the current
// firmware's schema doesn't recognize, or a value that fails the schema's
// declared value_type, is logged and skipped - never a reason to fail the
// whole registration, matching the firmware's own MQTT config handler's
// "unknown key" behavior.
func (u *usecase) syncReportedConfig(ctx context.Context, node *domainmodels.Node, reported map[string]string) error {
	const tag = "node/messaging_callback/syncReportedConfig"

	if len(reported) == 0 {
		return nil
	}

	params, err := u.firmwareConfigParams.ReadByFirmwareId(ctx, node.FirmwareId)
	if err != nil {
		return err
	}

	valueTypes := make(map[string]string, len(params))
	for _, param := range params {
		valueTypes[param.Key] = param.ValueType
	}

	for key, value := range reported {
		valueType, known := valueTypes[key]
		if !known {
			u.logger.Warn(ctx, tag, "registration reported a config key not in the node's current firmware schema", domainmodels.LoggerMeta{
				"device_id":   node.DeviceId,
				"node_id":     node.Id,
				"firmware_id": node.FirmwareId,
				"key":         key,
			})
			continue
		}

		if err := applicationshared.ValidateConfigValue(value, valueType); err != nil {
			u.logger.Warn(ctx, tag, "registration reported an invalid config value for its type, skipping", domainmodels.LoggerMeta{
				"err":        err,
				"device_id":  node.DeviceId,
				"node_id":    node.Id,
				"key":        key,
				"value_type": valueType,
			})
			continue
		}

		if err := u.nodeConfigValues.Upsert(ctx, node.Id, node.FirmwareId, key, value, nil); err != nil {
			return err
		}
	}

	return nil
}
