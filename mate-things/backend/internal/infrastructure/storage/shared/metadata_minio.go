package infrastructurestorageshared

import (
	"strings"

	"github.com/minio/minio-go/v7"
)

const ChecksumMetadataKey = "checksum-sha256"

func ChecksumFromMinioMetadata(info minio.ObjectInfo) string {
	if info.UserMetadata != nil {
		for key, value := range info.UserMetadata {
			if strings.EqualFold(key, ChecksumMetadataKey) {
				return value
			}
		}
	}

	for key, values := range info.Metadata {
		if strings.EqualFold(key, "X-Amz-Meta-"+ChecksumMetadataKey) || strings.EqualFold(key, ChecksumMetadataKey) {
			if len(values) > 0 {
				return values[0]
			}
		}
	}

	return ""
}
