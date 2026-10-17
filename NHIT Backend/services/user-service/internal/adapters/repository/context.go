package repository

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/jackc/pgx/v5"
)

type txKey struct{}

// getQueries returns the sqlc.Queries instance from the context if a transaction is present,
// otherwise it returns the provided default queries.
func getQueries(ctx context.Context, defaultQueries *sqlc.Queries) *sqlc.Queries {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return defaultQueries.WithTx(tx)
	}
	return defaultQueries
}

// withTxContext returns a new context with the provided transaction.
func withTxContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
