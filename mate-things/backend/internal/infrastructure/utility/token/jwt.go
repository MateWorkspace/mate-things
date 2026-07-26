package infrastructureutilitytoken

import (
	"errors"
	"time"

	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtImpl struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
	now             func() time.Time
}

type accessClaims struct {
	domainmodels.TokenClaimsAccess
	jwt.RegisteredClaims
}

type refreshClaims struct {
	domainmodels.TokenClaimsRefresh
	jwt.RegisteredClaims
}

func NewJwtImpl(
	accessSecret string,
	refreshSecret string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) domaincontractsutility.Token {
	return &jwtImpl{
		accessSecret:    []byte(accessSecret),
		refreshSecret:   []byte(refreshSecret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		now:             time.Now,
	}
}

func (j *jwtImpl) GenerateAccess(claims *domainmodels.TokenClaimsAccess) (accessToken string, err error) {
	if claims == nil {
		return "", domainmodels.NewError("access token claims are required", domainmodels.ErrTypeBadArgs, nil)
	}
	if len(j.accessSecret) == 0 {
		return "", domainmodels.NewError("access token secret is not configured", domainmodels.ErrTypeFailure, nil)
	}

	now := j.now().UTC()
	jwtClaims := accessClaims{
		TokenClaimsAccess: *claims,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserId.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessDuration)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims).SignedString(j.accessSecret)
	if err != nil {
		return "", domainmodels.NewError("failed to generate access token", domainmodels.ErrTypeFailure, err)
	}

	return token, nil
}

func (j *jwtImpl) ValidateAccess(accessToken string) (claims *domainmodels.TokenClaimsAccess, err error) {
	if accessToken == "" {
		return nil, domainmodels.NewError("access token is required", domainmodels.ErrTypeBadArgs, nil)
	}
	if len(j.accessSecret) == 0 {
		return nil, domainmodels.NewError("access token secret is not configured", domainmodels.ErrTypeFailure, nil)
	}

	jwtClaims := &accessClaims{}
	token, err := jwt.ParseWithClaims(accessToken, jwtClaims, j.keyFunc(j.accessSecret))
	if err != nil {
		return nil, mapTokenError("failed to validate access token", err)
	}
	if !token.Valid {
		return nil, domainmodels.NewError("access token is invalid", domainmodels.ErrTypeTokenInvalid, nil)
	}

	userId, err := parseClaimUserId(jwtClaims.UserId, jwtClaims.Subject)
	if err != nil {
		return nil, domainmodels.NewError("access token user id is invalid", domainmodels.ErrTypeTokenInvalid, err)
	}
	jwtClaims.UserId = userId

	return &jwtClaims.TokenClaimsAccess, nil
}

func (j *jwtImpl) GenerateRefresh(claims *domainmodels.TokenClaimsRefresh) (refreshToken string, err error) {
	if claims == nil {
		return "", domainmodels.NewError("refresh token claims are required", domainmodels.ErrTypeBadArgs, nil)
	}
	if len(j.refreshSecret) == 0 {
		return "", domainmodels.NewError("refresh token secret is not configured", domainmodels.ErrTypeFailure, nil)
	}

	now := j.now().UTC()
	jwtClaims := refreshClaims{
		TokenClaimsRefresh: *claims,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserId.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshDuration)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims).SignedString(j.refreshSecret)
	if err != nil {
		return "", domainmodels.NewError("failed to generate refresh token", domainmodels.ErrTypeFailure, err)
	}

	return token, nil
}

func (j *jwtImpl) ValidateRefresh(refreshToken string) (claims *domainmodels.TokenClaimsRefresh, err error) {
	if refreshToken == "" {
		return nil, domainmodels.NewError("refresh token is required", domainmodels.ErrTypeBadArgs, nil)
	}
	if len(j.refreshSecret) == 0 {
		return nil, domainmodels.NewError("refresh token secret is not configured", domainmodels.ErrTypeFailure, nil)
	}

	jwtClaims := &refreshClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, jwtClaims, j.keyFunc(j.refreshSecret))
	if err != nil {
		return nil, mapTokenError("failed to validate refresh token", err)
	}
	if !token.Valid {
		return nil, domainmodels.NewError("refresh token is invalid", domainmodels.ErrTypeTokenInvalid, nil)
	}

	userId, err := parseClaimUserId(jwtClaims.UserId, jwtClaims.Subject)
	if err != nil {
		return nil, domainmodels.NewError("refresh token user id is invalid", domainmodels.ErrTypeTokenInvalid, err)
	}
	jwtClaims.UserId = userId

	return &jwtClaims.TokenClaimsRefresh, nil
}

func (j *jwtImpl) keyFunc(secret []byte) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		return secret, nil
	}
}

func parseClaimUserId(userId uuid.UUID, subject string) (uuid.UUID, error) {
	if userId != uuid.Nil {
		return userId, nil
	}

	return uuid.Parse(subject)
}

func mapTokenError(message string, err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return domainmodels.NewError(message, domainmodels.ErrTypeTokenExpired, err)
	case errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenUnverifiable),
		errors.Is(err, jwt.ErrTokenInvalidClaims),
		errors.Is(err, jwt.ErrTokenNotValidYet):
		return domainmodels.NewError(message, domainmodels.ErrTypeTokenInvalid, err)
	default:
		return domainmodels.NewError(message, domainmodels.ErrTypeTokenInvalid, err)
	}
}
