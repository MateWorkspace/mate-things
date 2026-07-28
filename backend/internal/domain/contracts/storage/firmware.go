package domaincontractsstorage

import (
	"context"
	"io"
	"time"
)

type Firmware interface {
	Store(
		ctx context.Context,
		name string,
		content io.Reader,
	) (path string, size int32, checksum string, err error)

	Presign(
		ctx context.Context,
		name string,
		downloadFilename string,
	) (url string, expiresAt time.Time, err error)

	Stat(
		ctx context.Context,
		name string,
	) (path string, size int32, checksum string, err error)

	Delete(
		ctx context.Context,
		name string,
	) (err error)
}
