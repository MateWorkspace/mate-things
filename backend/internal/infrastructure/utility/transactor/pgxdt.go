package infrastructurerepositorytransactor

import (
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
)

func NewPgxdtImpl(dt pgxdt.Transactor) domaincontractsutility.Transactor {
	return dt
}
