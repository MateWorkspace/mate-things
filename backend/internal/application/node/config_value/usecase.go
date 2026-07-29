package applicationnodeconfigvalue

import (
	"context"
	"strconv"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	repository          domaincontractsrepository.NodeConfigValue
	parameterRepository domaincontractsrepository.FirmwareConfigParameter
	node                domainusecasesrepocache.Node
	publisher           domaincontractsnode.Publish
	logger              domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	repository domaincontractsrepository.NodeConfigValue,
	parameterRepository domaincontractsrepository.FirmwareConfigParameter,
	node domainusecasesrepocache.Node,
	publisher domaincontractsnode.Publish,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.ConfigValue {
	return &usecase{
		repository:          repository,
		parameterRepository: parameterRepository,
		node:                node,
		publisher:           publisher,
		logger:              logger,
	}
}

// ReadByNodeId returns only the values whose key still exists in the node's
// CURRENT firmware's config schema. Stored rows stay keyed to the firmware they
// were set under, so a node that was reassigned to another firmware keeps its
// old rows as history, but they are hidden here to keep the read surface
// identical to what SetByNodeId will accept.
func (u *usecase) ReadByNodeId(
	ctx context.Context,
	request domainusecasesnode.ReadConfigValuesByNodeIdRequest,
) ([]domainmodels.NodeConfigValue, error) {
	const tag = "node/config_value/ReadByNodeId"

	node, err := u.node.ReadById(ctx, request.NodeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":     err,
			"node_id": request.NodeId,
		})
		return nil, err
	}

	values, err := u.repository.ReadByNodeId(ctx, request.NodeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node config values", domainmodels.LoggerMeta{
			"err":     err,
			"node_id": request.NodeId,
		})
		return nil, err
	}

	params, err := u.parameterRepository.ReadByFirmwareId(ctx, node.FirmwareId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware config parameters", domainmodels.LoggerMeta{
			"err":         err,
			"firmware_id": node.FirmwareId,
		})
		return nil, err
	}

	currentKeys := make(map[string]bool, len(params))
	for _, param := range params {
		currentKeys[param.Key] = true
	}

	current := make([]domainmodels.NodeConfigValue, 0, len(values))
	for _, value := range values {
		if currentKeys[value.Key] {
			current = append(current, value)
		}
	}

	return current, nil
}

func (u *usecase) SetByNodeId(ctx context.Context, request domainusecasesnode.SetConfigValueRequest) error {
	const tag = "node/config_value/SetByNodeId"

	node, err := u.node.ReadById(ctx, request.NodeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":     err,
			"node_id": request.NodeId,
		})
		return err
	}

	params, err := u.parameterRepository.ReadByFirmwareId(ctx, node.FirmwareId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware config parameters", domainmodels.LoggerMeta{
			"err":         err,
			"firmware_id": node.FirmwareId,
		})
		return err
	}

	var valueType string
	found := false
	for _, param := range params {
		if param.Key == request.Key {
			valueType = param.ValueType
			found = true
			break
		}
	}
	if !found {
		return domainmodels.NewError("config key does not exist on the node's current firmware", domainmodels.ErrTypeNotFound, nil)
	}

	if err := validateConfigValue(request.Value, valueType); err != nil {
		return err
	}

	if err := u.repository.Upsert(ctx, request.NodeId, node.FirmwareId, request.Key, request.Value, request.ActorId); err != nil {
		u.logger.Error(ctx, tag, "failed to upsert node config value", domainmodels.LoggerMeta{
			"err":     err,
			"node_id": request.NodeId,
			"key":     request.Key,
		})
		return err
	}

	if err := u.publisher.Config(ctx, node.DeviceId, request.Key, request.Value); err != nil {
		u.logger.Error(ctx, tag, "failed to publish config value", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": node.DeviceId,
			"key":       request.Key,
		})
		return err
	}

	return nil
}

func validateConfigValue(value string, valueType string) error {
	switch valueType {
	case "uint32":
		if _, err := strconv.ParseUint(value, 10, 32); err != nil {
			return domainmodels.NewError("config value must be a valid uint32", domainmodels.ErrTypeValidation, err)
		}
	case "bool":
		if value != "true" && value != "false" {
			return domainmodels.NewError("config value must be \"true\" or \"false\"", domainmodels.ErrTypeValidation, nil)
		}
	case "string":
		// any string value is acceptable
	default:
		return domainmodels.NewError("unrecognized config value_type", domainmodels.ErrTypeFailure, nil)
	}

	return nil
}
