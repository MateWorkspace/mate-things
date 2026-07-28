package infrastructurerepositoryshared

import (
	"github.com/Masterminds/squirrel"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
)

type BasePostgres struct {
	Dt   pgxdt.Pgxdt
	SqrD *squirrel.StatementBuilderType
	SqrQ *squirrel.StatementBuilderType
}
