package applicationnodeconfigparameter

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

var validValueTypes = map[string]bool{
	"string": true,
	"uint32": true,
	"bool":   true,
}

type usecase struct {
	repository domaincontractsrepository.FirmwareConfigParameter
	logger     domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	repository domaincontractsrepository.FirmwareConfigParameter,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.ConfigParameter {
	return &usecase{
		repository: repository,
		logger:     logger,
	}
}

func (u *usecase) ReplaceForFirmware(ctx context.Context, request domainusecasesnode.ReplaceConfigParametersRequest) error {
	const tag = "node/config_parameter/ReplaceForFirmware"

	params := make([]domaincontractsrepository.FirmwareConfigParameterInput, 0, len(request.Parameters))
	for _, param := range request.Parameters {
		if !validValueTypes[param.ValueType] {
			return domainmodels.NewError("config parameter value_type must be one of string, uint32, bool", domainmodels.ErrTypeValidation, nil)
		}
		if param.Key == "" {
			return domainmodels.NewError("config parameter key is required", domainmodels.ErrTypeValidation, nil)
		}
		params = append(params, domaincontractsrepository.FirmwareConfigParameterInput{
			Key:       param.Key,
			ValueType: param.ValueType,
		})
	}

	if err := u.repository.ReplaceForFirmwareId(ctx, request.FirmwareId, params, request.ActorId); err != nil {
		u.logger.Error(ctx, tag, "failed to replace firmware config parameters", domainmodels.LoggerMeta{
			"err":         err,
			"firmware_id": request.FirmwareId,
		})
		return err
	}

	return nil
}

func (u *usecase) ReadByFirmwareId(
	ctx context.Context,
	request domainusecasesnode.ReadConfigParametersByFirmwareIdRequest,
) ([]domainmodels.FirmwareConfigParameter, error) {
	const tag = "node/config_parameter/ReadByFirmwareId"

	params, err := u.repository.ReadByFirmwareId(ctx, request.FirmwareId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware config parameters", domainmodels.LoggerMeta{
			"err":         err,
			"firmware_id": request.FirmwareId,
		})
		return nil, err
	}

	return params, nil
}
