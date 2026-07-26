package domaincontractsutility

import domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"

type Token interface {
	GenerateAccess(claims *domainmodels.TokenClaimsAccess) (accessToken string, err error)
	ValidateAccess(accessToken string) (claims *domainmodels.TokenClaimsAccess, err error)
	GenerateRefresh(claims *domainmodels.TokenClaimsRefresh) (refreshToken string, err error)
	ValidateRefresh(refreshToken string) (claims *domainmodels.TokenClaimsRefresh, err error)
}
