package infrastructurestorageshared

import (
	"context"
	"errors"
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/minio/minio-go/v7"
)

func MapMinioError(message string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return domainmodels.NewError(message, domainmodels.ErrTypeTimeout, err)
	}

	resp := minio.ToErrorResponse(err)
	switch resp.Code {
	case "NoSuchBucket", "NoSuchKey", "NoSuchObject", "NotFound":
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch":
		return domainmodels.NewError(message, domainmodels.ErrTypeUnauthorized, err)
	case "InvalidArgument", "InvalidBucketName", "InvalidObjectName", "MalformedXML":
		return domainmodels.NewError(message, domainmodels.ErrTypeValidation, err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	case http.StatusUnauthorized, http.StatusForbidden:
		return domainmodels.NewError(message, domainmodels.ErrTypeUnauthorized, err)
	case http.StatusBadRequest:
		return domainmodels.NewError(message, domainmodels.ErrTypeValidation, err)
	}

	return domainmodels.NewError(message, domainmodels.ErrTypeFailure, err)
}
