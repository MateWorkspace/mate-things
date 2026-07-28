package presentationhttpmiddleware

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

func Permission(requiredPermissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if len(requiredPermissions) == 0 {
				return next(c)
			}

			claims := presentationhttputils.AccessClaims(c.Request().Context())
			if claims == nil {
				return presentationhttputils.Error(c, domainmodels.NewError("authenticated user is required", domainmodels.ErrTypeUnauthorized, nil))
			}

			granted := make(map[string]struct{}, len(claims.Permissions))
			for _, permission := range claims.Permissions {
				granted[permission] = struct{}{}
			}

			for _, permission := range requiredPermissions {
				if _, ok := granted[permission]; !ok {
					return presentationhttputils.Error(c, domainmodels.NewError("permission is denied", domainmodels.ErrTypeUnauthorized, nil))
				}
			}

			return next(c)
		}
	}
}
