package applicationnodefirmwaremanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	firmware        domainusecasesrepocache.Firmware
	node            domainusecasesrepocache.Node
	storage         domaincontractsstorage.Firmware
	configParameter domainusecasesnode.ConfigParameter
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	firmware domainusecasesrepocache.Firmware,
	node domainusecasesrepocache.Node,
	storage domaincontractsstorage.Firmware,
	configParameter domainusecasesnode.ConfigParameter,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.FirmwareManagement {
	return &usecase{
		firmware:        firmware,
		node:            node,
		storage:         storage,
		configParameter: configParameter,
		logger:          logger,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	request domainusecasesnode.CreateFirmwareRequest,
) (domainusecasesnode.CreateFirmwareResult, error) {
	const tag = "node/firmware_management/Create"

	name, err := applicationshared.RequiredFirmwareName(request.Name, "name")
	if err != nil {
		return domainusecasesnode.CreateFirmwareResult{}, err
	}
	content, err := applicationshared.RequiredFirmwareContent(request.Content, "file")
	if err != nil {
		return domainusecasesnode.CreateFirmwareResult{}, err
	}

	path, size, checksum, err := u.storage.Store(ctx, name, content)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to store firmware binary", domainmodels.LoggerMeta{
			"err":           err,
			"name":          name,
			"node_class_id": request.NodeClassId,
			"created_by":    request.CreatedBy,
		})
		return domainusecasesnode.CreateFirmwareResult{}, err
	}

	id, err := u.firmware.Create(ctx, request.NodeClassId, name, size, checksum, path, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create firmware", domainmodels.LoggerMeta{
			"err":           err,
			"name":          request.Name,
			"node_class_id": request.NodeClassId,
			"binary_path":   path,
			"created_by":    request.CreatedBy,
		})
		if cleanupErr := u.storage.Delete(ctx, path); cleanupErr != nil {
			u.logger.Error(ctx, tag, "failed to cleanup stored firmware binary", domainmodels.LoggerMeta{
				"err":         cleanupErr,
				"binary_path": path,
			})
		}
		return domainusecasesnode.CreateFirmwareResult{}, err
	}

	if len(request.ConfigSchema) > 0 {
		if err := u.configParameter.ReplaceForFirmware(ctx, domainusecasesnode.ReplaceConfigParametersRequest{
			FirmwareId: id,
			Parameters: request.ConfigSchema,
			ActorId:    request.CreatedBy,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to ingest firmware config schema", domainmodels.LoggerMeta{
				"err":         err,
				"firmware_id": id,
			})
			return domainusecasesnode.CreateFirmwareResult{}, err
		}
	}

	return domainusecasesnode.CreateFirmwareResult{
		Id:         id,
		Size:       size,
		Checksum:   checksum,
		BinaryPath: path,
	}, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesnode.ReadFirmwareByIdRequest,
) (*domainmodels.Firmware, error) {
	const tag = "node/firmware_management/ReadById"

	firmware, err := u.firmware.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return firmware, nil
}

func (u *usecase) ReadByName(
	ctx context.Context,
	request domainusecasesnode.ReadFirmwareByNameRequest,
) (*domainmodels.Firmware, error) {
	const tag = "node/firmware_management/ReadByName"

	firmware, err := u.firmware.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return firmware, nil
}

func (u *usecase) ReadAvailableByNodeId(
	ctx context.Context,
	request domainusecasesnode.ReadAvailableFirmwaresByNodeIdRequest,
) ([]domainmodels.Firmware, int, error) {
	const tag = "node/firmware_management/ReadAvailableByNodeId"

	node, err := u.node.ReadById(ctx, request.NodeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":     err,
			"node_id": request.NodeId,
		})
		return nil, 0, err
	}

	firmwares, total, err := u.firmware.ReadByNodeClassIdAndPagination(
		ctx,
		node.NodeClassId,
		request.Page,
		request.Limit,
		request.Search,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read available firmwares", domainmodels.LoggerMeta{
			"err":           err,
			"node_id":       request.NodeId,
			"node_class_id": node.NodeClassId,
			"page":          request.Page,
			"limit":         request.Limit,
		})
		return nil, 0, err
	}

	return firmwares, total, nil
}

func (u *usecase) ReadByNodeClassIdAndPagination(
	ctx context.Context,
	request domainusecasesnode.ReadFirmwaresByNodeClassIdAndPaginationRequest,
) ([]domainmodels.Firmware, int, error) {
	const tag = "node/firmware_management/ReadByNodeClassIdAndPagination"

	firmwares, total, err := u.firmware.ReadByNodeClassIdAndPagination(
		ctx,
		request.NodeClassId,
		request.Page,
		request.Limit,
		request.Search,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmwares", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"page":          request.Page,
			"limit":         request.Limit,
		})
		return nil, 0, err
	}

	return firmwares, total, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesnode.ReadFirmwaresByPaginationRequest,
) ([]domainmodels.Firmware, int, error) {
	const tag = "node/firmware_management/ReadByPagination"

	firmwares, total, err := u.firmware.ReadByPagination(ctx, request.Page, request.Limit, request.Search, request.NodeClassId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmwares", domainmodels.LoggerMeta{
			"err":           err,
			"page":          request.Page,
			"limit":         request.Limit,
			"node_class_id": request.NodeClassId,
		})
		return nil, 0, err
	}

	return firmwares, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesnode.UpdateFirmwareRequest) error {
	const tag = "node/firmware_management/UpdateById"

	name, err := applicationshared.OptionalFirmwareName(request.Name, "name")
	if err != nil {
		return err
	}

	if err := u.firmware.UpdateById(
		ctx,
		request.Id,
		request.NodeClassId,
		name,
		nil,
		nil,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update firmware", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) ReplaceBinaryById(
	ctx context.Context,
	request domainusecasesnode.ReplaceFirmwareBinaryByIdRequest,
) (domainusecasesnode.FirmwareBinaryStatResult, error) {
	const tag = "node/firmware_management/ReplaceBinaryById"

	firmware, err := u.firmware.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	content, err := applicationshared.RequiredFirmwareContent(request.Content, "file")
	if err != nil {
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	path, size, checksum, err := u.storage.Store(ctx, firmware.Name, content)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to store firmware binary", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	if err := u.firmware.UpdateById(ctx, request.Id, nil, nil, &size, &checksum, &path, nil, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update firmware binary metadata", domainmodels.LoggerMeta{
			"err":         err,
			"id":          request.Id,
			"binary_path": path,
			"updated_by":  request.UpdatedBy,
		})
		if cleanupErr := u.storage.Delete(ctx, path); cleanupErr != nil {
			u.logger.Error(ctx, tag, "failed to cleanup stored firmware binary", domainmodels.LoggerMeta{
				"err":         cleanupErr,
				"binary_path": path,
			})
		}
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	if len(request.ConfigSchema) > 0 {
		if err := u.configParameter.ReplaceForFirmware(ctx, domainusecasesnode.ReplaceConfigParametersRequest{
			FirmwareId: request.Id,
			Parameters: request.ConfigSchema,
			ActorId:    request.UpdatedBy,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to ingest firmware config schema", domainmodels.LoggerMeta{
				"err":         err,
				"firmware_id": request.Id,
			})
			return domainusecasesnode.FirmwareBinaryStatResult{}, err
		}
	}

	return domainusecasesnode.FirmwareBinaryStatResult{
		BinaryPath: path,
		Size:       size,
		Checksum:   checksum,
	}, nil
}

func (u *usecase) DownloadUrlById(
	ctx context.Context,
	request domainusecasesnode.DownloadFirmwareBinaryByIdRequest,
) (domainusecasesnode.FirmwareDownloadResult, error) {
	const tag = "node/firmware_management/DownloadUrlById"

	firmware, err := u.firmware.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.FirmwareDownloadResult{}, err
	}

	downloadUrl, expiresAt, err := u.storage.Presign(ctx, firmware.BinaryPath, firmware.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to presign firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"id":          request.Id,
			"binary_path": firmware.BinaryPath,
		})
		return domainusecasesnode.FirmwareDownloadResult{}, err
	}

	return domainusecasesnode.FirmwareDownloadResult{
		Firmware:    *firmware,
		DownloadUrl: downloadUrl,
		ExpiresAt:   expiresAt,
	}, nil
}

func (u *usecase) DownloadUrlByName(
	ctx context.Context,
	request domainusecasesnode.DownloadFirmwareBinaryByNameRequest,
) (domainusecasesnode.FirmwareDownloadResult, error) {
	const tag = "node/firmware_management/DownloadUrlByName"

	firmware, err := u.firmware.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return domainusecasesnode.FirmwareDownloadResult{}, err
	}

	downloadUrl, expiresAt, err := u.storage.Presign(ctx, firmware.BinaryPath, firmware.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to presign firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"name":        request.Name,
			"binary_path": firmware.BinaryPath,
		})
		return domainusecasesnode.FirmwareDownloadResult{}, err
	}

	return domainusecasesnode.FirmwareDownloadResult{
		Firmware:    *firmware,
		DownloadUrl: downloadUrl,
		ExpiresAt:   expiresAt,
	}, nil
}

func (u *usecase) StatBinaryByName(
	ctx context.Context,
	request domainusecasesnode.StatFirmwareBinaryByNameRequest,
) (domainusecasesnode.FirmwareBinaryStatResult, error) {
	const tag = "node/firmware_management/StatBinaryByName"

	firmware, err := u.firmware.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	path, size, checksum, err := u.storage.Stat(ctx, firmware.BinaryPath)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to stat firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"name":        request.Name,
			"binary_path": firmware.BinaryPath,
		})
		return domainusecasesnode.FirmwareBinaryStatResult{}, err
	}

	return domainusecasesnode.FirmwareBinaryStatResult{
		BinaryPath: path,
		Size:       size,
		Checksum:   checksum,
	}, nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesnode.DeleteFirmwareRequest) error {
	const tag = "node/firmware_management/DeleteById"

	firmware, err := u.firmware.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return err
	}

	if err := u.firmware.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete firmware", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}
	if err := u.storage.Delete(ctx, firmware.BinaryPath); err != nil {
		u.logger.Error(ctx, tag, "failed to delete firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"id":          request.Id,
			"binary_path": firmware.BinaryPath,
			"deleted_by":  request.DeletedBy,
		})
		return err
	}

	return nil
}
