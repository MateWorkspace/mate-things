package presentationhttputils

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type contextKey string

const accessClaimsContextKey contextKey = "http_access_claims"

func InjectAccessClaims(ctx context.Context, claims *domainmodels.TokenClaimsAccess) context.Context {
	return context.WithValue(ctx, accessClaimsContextKey, claims)
}

func AccessClaims(ctx context.Context) *domainmodels.TokenClaimsAccess {
	claims, _ := ctx.Value(accessClaimsContextKey).(*domainmodels.TokenClaimsAccess)
	return claims
}

func ActorId(c *echo.Context) *uuid.UUID {
	claims := AccessClaims(c.Request().Context())
	if claims == nil || claims.UserId == uuid.Nil {
		return nil
	}

	userId := claims.UserId
	return &userId
}

func RequiredActorId(c *echo.Context) (uuid.UUID, error) {
	actorId := ActorId(c)
	if actorId == nil || *actorId == uuid.Nil {
		return uuid.Nil, domainmodels.NewError("authenticated user is required", domainmodels.ErrTypeUnauthorized, nil)
	}

	return *actorId, nil
}
