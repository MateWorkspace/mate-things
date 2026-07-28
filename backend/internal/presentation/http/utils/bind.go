package presentationhttputils

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/labstack/echo/v5"
)

func Bind(c *echo.Context, value any) error {
	if err := c.Bind(value); err != nil {
		return Error(c, domainmodels.NewError(
			"request body is invalid",
			domainmodels.ErrTypeValidation,
			err,
		), "The request body could not be processed. Please check your input.")
	}

	return nil
}
