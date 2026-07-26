package infrastructurerepositoryshared

import (
	"github.com/ABA-Developer/nusapala-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
)

type BasePostgres struct {
	Dt   pgxdt.Pgxdt
	SqrD *squirrel.StatementBuilderType
	SqrQ *squirrel.StatementBuilderType
}
