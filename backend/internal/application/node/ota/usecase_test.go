package applicationnodeota

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

func TestDispatchByNodeIdResolvesAuthoritativeCompatibleFirmwareBeforePublishing(t *testing.T) {
	nodeID := uuid.New()
	nodeClassID := uuid.New()
	firmwareID := uuid.New()
	node := &otaNodeRepository{
		node: &domainmodels.Node{
			Id:          nodeID,
			NodeClassId: nodeClassID,
			DeviceId:    "AC276E5E030C",
		},
	}
	firmware := &otaFirmwareRepository{
		firmware: &domainmodels.Firmware{
			Id:          firmwareID,
			NodeClassId: nodeClassID,
			Name:        "freezer-v2",
			Size:        2048,
			Checksum:    "authoritative-checksum",
			BinaryPath:  "firmwares/freezer-v2.bin",
		},
	}
	storage := &otaFirmwareStorage{url: "https://storage.example/presigned"}
	publisher := &otaPublisher{}
	usecase := NewUsecaseImpl(
		node,
		firmware,
		storage,
		publisher,
		&otaLogger{},
	)

	err := usecase.DispatchByNodeId(context.Background(), domainusecasesnode.DispatchOtaByNodeIdRequest{
		NodeId:     nodeID,
		FirmwareId: firmwareID,
	})

	if err != nil {
		t.Fatalf("DispatchByNodeId() error = %v, want nil", err)
	}
	if storage.presignCalls != 1 {
		t.Fatalf("storage Presign() calls = %d, want 1", storage.presignCalls)
	}
	if storage.path != firmware.firmware.BinaryPath || storage.downloadFilename != firmware.firmware.Name {
		t.Fatalf(
			"storage Presign() = (%q, %q), want (%q, %q)",
			storage.path,
			storage.downloadFilename,
			firmware.firmware.BinaryPath,
			firmware.firmware.Name,
		)
	}
	if publisher.otaCalls != 1 {
		t.Fatalf("publisher Ota() calls = %d, want 1", publisher.otaCalls)
	}
	if publisher.deviceID != node.node.DeviceId ||
		publisher.url != storage.url ||
		publisher.size != firmware.firmware.Size ||
		publisher.checksum != firmware.firmware.Checksum {
		t.Fatalf(
			"publisher Ota() = (%q, %q, %d, %q), want (%q, %q, %d, %q)",
			publisher.deviceID,
			publisher.url,
			publisher.size,
			publisher.checksum,
			node.node.DeviceId,
			storage.url,
			firmware.firmware.Size,
			firmware.firmware.Checksum,
		)
	}
}

func TestDispatchByNodeIdRejectsIncompatibleFirmwareBeforePresignOrPublish(t *testing.T) {
	nodeID := uuid.New()
	firmwareID := uuid.New()
	storage := &otaFirmwareStorage{url: "https://storage.example/presigned"}
	publisher := &otaPublisher{}
	usecase := NewUsecaseImpl(
		&otaNodeRepository{
			node: &domainmodels.Node{
				Id:          nodeID,
				NodeClassId: uuid.New(),
				DeviceId:    "AC276E5E030C",
			},
		},
		&otaFirmwareRepository{
			firmware: &domainmodels.Firmware{
				Id:          firmwareID,
				NodeClassId: uuid.New(),
				Name:        "wrong-class-v2",
				BinaryPath:  "firmwares/wrong-class-v2.bin",
			},
		},
		storage,
		publisher,
		&otaLogger{},
	)

	err := usecase.DispatchByNodeId(context.Background(), domainusecasesnode.DispatchOtaByNodeIdRequest{
		NodeId:     nodeID,
		FirmwareId: firmwareID,
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("DispatchByNodeId() error = %v, want validation error", err)
	}
	if storage.presignCalls != 0 {
		t.Fatalf("storage Presign() calls = %d, want 0", storage.presignCalls)
	}
	if publisher.otaCalls != 0 {
		t.Fatalf("publisher Ota() calls = %d, want 0", publisher.otaCalls)
	}
}

type otaNodeRepository struct {
	domainusecasesrepocache.Node
	node *domainmodels.Node
	err  error
}

func (r *otaNodeRepository) ReadById(context.Context, uuid.UUID) (*domainmodels.Node, error) {
	return r.node, r.err
}

type otaFirmwareRepository struct {
	domainusecasesrepocache.Firmware
	firmware *domainmodels.Firmware
	err      error
}

func (r *otaFirmwareRepository) ReadById(context.Context, uuid.UUID) (*domainmodels.Firmware, error) {
	return r.firmware, r.err
}

func (r *otaFirmwareRepository) UpdateById(
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
	return nil
}

type otaFirmwareStorage struct {
	domaincontractsstorage.Firmware
	url              string
	path             string
	downloadFilename string
	presignCalls     int
}

func (s *otaFirmwareStorage) Presign(
	_ context.Context,
	path string,
	downloadFilename string,
) (string, time.Time, error) {
	s.presignCalls++
	s.path = path
	s.downloadFilename = downloadFilename
	return s.url, time.Now().Add(time.Minute), nil
}

type otaPublisher struct {
	domaincontractsnode.Publish
	otaCalls int
	deviceID string
	url      string
	size     int32
	checksum string
}

func (p *otaPublisher) Ota(
	_ context.Context,
	deviceID string,
	url string,
	size int32,
	checksum string,
) error {
	p.otaCalls++
	p.deviceID = deviceID
	p.url = url
	p.size = size
	p.checksum = checksum
	return nil
}

type otaLogger struct {
	domaincontractslogger.Leveled
}
