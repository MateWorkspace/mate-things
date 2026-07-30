package applicationnodefirmwaremanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

func TestDeleteByIdRejectsSpoofedConfirmationBeforeDeleting(t *testing.T) {
	firmwareID := uuid.New()
	firmware := &recordingFirmware{
		readFirmware: &domainmodels.Firmware{
			Id:         firmwareID,
			Name:       "authoritative-firmware",
			BinaryPath: "firmwares/authoritative-firmware.bin",
		},
	}
	storage := &recordingFirmwareStorage{}
	usecase := NewUsecaseImpl(
		firmware,
		&recordingNode{},
		storage,
		&recordingConfigParameter{},
		&recordingLogger{},
	)

	err := usecase.DeleteById(context.Background(), domainusecasesnode.DeleteFirmwareRequest{
		Id:           firmwareID,
		ExpectedName: "spoofed-firmware",
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("DeleteById() error = %v, want validation error", err)
	}
	if firmware.deleteCalls != 0 {
		t.Fatalf("firmware DeleteById() calls = %d, want 0", firmware.deleteCalls)
	}
	if storage.deleteCalls != 0 {
		t.Fatalf("storage Delete() calls = %d, want 0", storage.deleteCalls)
	}
}

func TestReplaceBinaryByIdClearsExistingConfigSchemaWhenRequestIsEmpty(t *testing.T) {
	firmwareID := uuid.New()
	firmware := &recordingFirmware{
		readFirmware: &domainmodels.Firmware{
			Id:   firmwareID,
			Name: "firmware-v2",
		},
	}
	config := &recordingConfigParameter{}
	usecase := NewUsecaseImpl(
		firmware,
		&recordingNode{},
		&recordingFirmwareStorage{
			storePath:     "firmwares/firmware-v2.bin",
			storeSize:     17,
			storeChecksum: "checksum-v2",
		},
		config,
		&recordingLogger{},
	)

	_, err := usecase.ReplaceBinaryById(context.Background(), domainusecasesnode.ReplaceFirmwareBinaryByIdRequest{
		Id:           firmwareID,
		Content:      bytes.NewReader([]byte{0xE9, 0x01, 0x02}),
		ConfigSchema: []domainusecasesnode.ConfigParameterInput{},
	})

	if err != nil {
		t.Fatalf("ReplaceBinaryById() error = %v, want nil", err)
	}
	if config.replaceCalls != 1 {
		t.Fatalf("ReplaceForFirmware() calls = %d, want 1", config.replaceCalls)
	}
	if config.replaceRequest.FirmwareId != firmwareID {
		t.Fatalf("ReplaceForFirmware() firmware ID = %s, want %s", config.replaceRequest.FirmwareId, firmwareID)
	}
	if len(config.replaceRequest.Parameters) != 0 {
		t.Fatalf("ReplaceForFirmware() parameters = %#v, want empty", config.replaceRequest.Parameters)
	}
}

type recordingFirmware struct {
	domainusecasesrepocache.Firmware
	readFirmware *domainmodels.Firmware
	readErr      error
	updateCalls  int
	deleteCalls  int
}

func (r *recordingFirmware) ReadById(context.Context, uuid.UUID) (*domainmodels.Firmware, error) {
	return r.readFirmware, r.readErr
}

func (r *recordingFirmware) UpdateById(
	context.Context,
	uuid.UUID,
	*uuid.UUID,
	*string,
	*int32,
	*string,
	*string,
	*json.RawMessage,
	*uuid.UUID,
) error {
	r.updateCalls++
	return nil
}

func (r *recordingFirmware) DeleteById(context.Context, uuid.UUID, *uuid.UUID) error {
	r.deleteCalls++
	return nil
}

type recordingNode struct {
	domainusecasesrepocache.Node
}

type recordingFirmwareStorage struct {
	domaincontractsstorage.Firmware
	storePath     string
	storeSize     int32
	storeChecksum string
	storeCalls    int
	deleteCalls   int
}

func (r *recordingFirmwareStorage) Store(context.Context, string, io.Reader) (string, int32, string, error) {
	r.storeCalls++
	return r.storePath, r.storeSize, r.storeChecksum, nil
}

func (r *recordingFirmwareStorage) Delete(context.Context, string) error {
	r.deleteCalls++
	return nil
}

type recordingConfigParameter struct {
	domainusecasesnode.ConfigParameter
	validateCalls  int
	replaceCalls   int
	replaceRequest domainusecasesnode.ReplaceConfigParametersRequest
}

func (r *recordingConfigParameter) ValidateSchema([]domainusecasesnode.ConfigParameterInput) error {
	r.validateCalls++
	return nil
}

func (r *recordingConfigParameter) ReplaceForFirmware(
	_ context.Context,
	request domainusecasesnode.ReplaceConfigParametersRequest,
) error {
	r.replaceCalls++
	r.replaceRequest = request
	return nil
}

type recordingLogger struct {
	domaincontractslogger.Leveled
}
