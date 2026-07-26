package infrastructurerepositorytransactor

import (
	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	"github.com/ABA-Developer/nusapala-things/backend/pkg/pgxdt"
)

func NewPgxdtImpl(dt pgxdt.Transactor) domaincontractsutility.Transactor {
	return dt
}
