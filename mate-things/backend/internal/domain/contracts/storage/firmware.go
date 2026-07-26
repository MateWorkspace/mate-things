package domaincontractsstorage

import (
	"context"
	"io"
)

type Firmware interface {
	Store(
		ctx context.Context,
		name string,
		content io.Reader,
	) (path string, size int32, checksum string, err error)

	Open(
		ctx context.Context,
		name string,
	) (Content io.ReadCloser, err error)

	Stat(
		ctx context.Context,
		name string,
	) (path string, size int32, checksum string, err error)

	Delete(
		ctx context.Context,
		name string,
	) (err error)
}
