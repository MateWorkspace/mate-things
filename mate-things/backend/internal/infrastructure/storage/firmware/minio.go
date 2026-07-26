package infrastructurestoragefirmware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"math"
	"os"
	"strings"

	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurestorageshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/storage/shared"
	"github.com/minio/minio-go/v7"
)

const contentTypeFirmware = "application/octet-stream"

type minioImpl struct {
	client *minio.Client
	bucket string
}

func NewMinioImpl(client *minio.Client, bucket string) domaincontractsstorage.Firmware {
	return &minioImpl{
		client: client,
		bucket: bucket,
	}
}

func (m *minioImpl) Store(
	ctx context.Context,
	name string,
	content io.Reader,
) (path string, size int32, checksum string, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", 0, "", domainmodels.NewError("firmware name is required", domainmodels.ErrTypeBadArgs, nil)
	}
	if content == nil {
		return "", 0, "", domainmodels.NewError("firmware content is required", domainmodels.ErrTypeBadArgs, nil)
	}

	file, size, checksum, err := spoolFirmware(content)
	if err != nil {
		return "", 0, "", err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	info, err := m.client.PutObject(
		ctx,
		m.bucket,
		name,
		file,
		int64(size),
		minio.PutObjectOptions{
			ContentType: contentTypeFirmware,
			UserMetadata: map[string]string{
				infrastructurestorageshared.ChecksumMetadataKey: checksum,
			},
		},
	)
	if err != nil {
		return "", 0, "", infrastructurestorageshared.MapMinioError("failed to store firmware", err)
	}

	path = info.Key
	if path == "" {
		path = name
	}

	return path, size, checksum, nil
}

func (m *minioImpl) Open(
	ctx context.Context,
	name string,
) (io.ReadCloser, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domainmodels.NewError("firmware name is required", domainmodels.ErrTypeBadArgs, nil)
	}

	content, err := m.client.GetObject(ctx, m.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, infrastructurestorageshared.MapMinioError("failed to open firmware", err)
	}
	if _, err := content.Stat(); err != nil {
		_ = content.Close()
		return nil, infrastructurestorageshared.MapMinioError("failed to open firmware", err)
	}

	return content, nil
}

func (m *minioImpl) Stat(
	ctx context.Context,
	name string,
) (path string, size int32, checksum string, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", 0, "", domainmodels.NewError("firmware name is required", domainmodels.ErrTypeBadArgs, nil)
	}

	info, err := m.client.StatObject(ctx, m.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return "", 0, "", infrastructurestorageshared.MapMinioError("failed to stat firmware", err)
	}

	if info.Size > math.MaxInt32 {
		return "", 0, "", domainmodels.NewError("firmware size exceeds int32 limit", domainmodels.ErrTypeValidation, nil)
	}

	checksum = infrastructurestorageshared.ChecksumFromMinioMetadata(info)
	if checksum == "" {
		return "", 0, "", domainmodels.NewError("firmware checksum metadata is missing", domainmodels.ErrTypeFailure, nil)
	}

	path = info.Key
	if path == "" {
		path = name
	}

	return path, int32(info.Size), checksum, nil
}

func (m *minioImpl) Delete(
	ctx context.Context,
	name string,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainmodels.NewError("firmware name is required", domainmodels.ErrTypeBadArgs, nil)
	}

	if _, _, _, err := m.Stat(ctx, name); err != nil {
		return err
	}

	if err := m.client.RemoveObject(ctx, m.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		return infrastructurestorageshared.MapMinioError("failed to delete firmware", err)
	}

	return nil
}

func spoolFirmware(content io.Reader) (*os.File, int32, string, error) {
	file, err := os.CreateTemp("", "nusapala-things-firmware-*")
	if err != nil {
		return nil, 0, "", domainmodels.NewError("failed to create firmware temporary file", domainmodels.ErrTypeFailure, err)
	}

	hasher := sha256.New()
	limit := int64(math.MaxInt32) + 1
	written, err := copyFirmware(file, hasher, io.LimitReader(content, limit))
	if err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, 0, "", err
	}
	if written > math.MaxInt32 {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, 0, "", domainmodels.NewError("firmware size exceeds int32 limit", domainmodels.ErrTypeValidation, nil)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, 0, "", domainmodels.NewError("failed to rewind firmware temporary file", domainmodels.ErrTypeFailure, err)
	}

	return file, int32(written), hex.EncodeToString(hasher.Sum(nil)), nil
}

func copyFirmware(file *os.File, hasher hash.Hash, content io.Reader) (int64, error) {
	written, err := io.Copy(io.MultiWriter(file, hasher), content)
	if err != nil {
		return written, domainmodels.NewError("failed to read firmware content", domainmodels.ErrTypeFailure, err)
	}

	return written, nil
}
