package presentationhttputils

import (
	"mime/multipart"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/labstack/echo/v5"
)

func RequiredFormFile(c *echo.Context, field string) (*multipart.FileHeader, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return nil, domainmodels.NewError(field+" file is required", domainmodels.ErrTypeValidation, err)
	}
	if file == nil || file.Size <= 0 {
		return nil, domainmodels.NewError(field+" file is required", domainmodels.ErrTypeValidation, nil)
	}

	return file, nil
}
