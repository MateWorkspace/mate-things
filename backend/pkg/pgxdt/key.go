package pgxdt

import (
	"context"
)

type txKey struct{}

func getTx(ctx context.Context) Pgxdt {
	if pgxdt, ok := ctx.Value(txKey{}).(Pgxdt); ok {
		return pgxdt
	}
	return nil
}

func putTx(ctx context.Context, inst Pgxdt) context.Context {
	return context.WithValue(ctx, txKey{}, inst)
}
