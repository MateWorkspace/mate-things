package presentationhttpmiddleware

import (
	"strings"

	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationhttputils "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func Auth(token domaincontractsutility.Token) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()

			accessToken := authorizationToken(req.Header.Get("Authorization"))
			if accessToken == "" {
				return presentationhttputils.Error(c, domainmodels.NewError("authorization is required", domainmodels.ErrTypeUnauthorized, nil))
			}

			claims, err := token.ValidateAccess(accessToken)
			if err != nil {
				return presentationhttputils.Error(c, err)
			}
			if claims == nil || claims.UserId == uuid.Nil {
				return presentationhttputils.Error(c, domainmodels.NewError("authenticated user is invalid", domainmodels.ErrTypeUnauthorized, nil))
			}

			ctx := presentationhttputils.InjectAccessClaims(req.Context(), claims)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

func authorizationToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 1 {
		return value
	}
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
