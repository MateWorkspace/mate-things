package applicationnodefirmwaremanagement

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	firmware domainusecasesrepocache.Firmware
	node     domainusecasesrepocache.Node
	storage  domaincontractsstorage.Firmware
	logger   domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	firmware domainusecasesrepocache.Firmware,
	node domainusecasesrepocache.Node,
	storage domaincontractsstorage.Firmware,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.FirmwareManagement {
	return &usecase{
		firmware: firmware,
		node:     node,
		storage:  storage,
		logger:   logger,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	request domainusecasesnode.CreateFirmwareRequest,
) (domainusecasesnode.CreateFirmwareResult, error) {
	const tag = "node/firmware_management/Create"

	path, size, checksum, err := u.storage.Store(ctx, request.Name, request.Content)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to store firmware binary", domainmodels.LoggerMeta{
			"err":           err,
			"name":          request.Name,
			"node_class_id": request.NodeClassId,
			"created_by":    request.CreatedBy,
		})
		return domainusecasesnode.CreateFirmwareResult{}, err
	}

	id, err := u.firmware.Create(ctx, request.NodeClassId, request.Name, size, checksum, path, request.CreatedBy)
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

	if err := u.firmware.UpdateById(
		ctx,
		request.Id,
		request.NodeClassId,
		request.Name,
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

	path, size, checksum, err := u.storage.Store(ctx, firmware.Name, request.Content)
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

	return domainusecasesnode.FirmwareBinaryStatResult{
		BinaryPath: path,
		Size:       size,
		Checksum:   checksum,
	}, nil
}

func (u *usecase) OpenBinaryById(
	ctx context.Context,
	request domainusecasesnode.OpenFirmwareBinaryByIdRequest,
) (domainusecasesnode.OpenFirmwareBinaryResult, error) {
	const tag = "node/firmware_management/OpenBinaryById"

	firmware, err := u.firmware.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.OpenFirmwareBinaryResult{}, err
	}

	content, err := u.storage.Open(ctx, firmware.BinaryPath)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to open firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"id":          request.Id,
			"binary_path": firmware.BinaryPath,
		})
		return domainusecasesnode.OpenFirmwareBinaryResult{}, err
	}

	return domainusecasesnode.OpenFirmwareBinaryResult{
		Firmware: *firmware,
		Content:  content,
	}, nil
}

func (u *usecase) OpenBinaryByName(
	ctx context.Context,
	request domainusecasesnode.OpenFirmwareBinaryByNameRequest,
) (domainusecasesnode.OpenFirmwareBinaryResult, error) {
	const tag = "node/firmware_management/OpenBinaryByName"

	firmware, err := u.firmware.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return domainusecasesnode.OpenFirmwareBinaryResult{}, err
	}

	content, err := u.storage.Open(ctx, firmware.BinaryPath)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to open firmware binary", domainmodels.LoggerMeta{
			"err":         err,
			"name":        request.Name,
			"binary_path": firmware.BinaryPath,
		})
		return domainusecasesnode.OpenFirmwareBinaryResult{}, err
	}

	return domainusecasesnode.OpenFirmwareBinaryResult{
		Firmware: *firmware,
		Content:  content,
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
