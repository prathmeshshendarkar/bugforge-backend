package postgres

import (
	"context"

	"bugforge-backend/internal/repository/interfaces"
)

// Tx executes fn inside a DB transaction. It provides a KanbanRepository
// backed by a TxExecutor to the function.
func (r *KanbanRepo) Tx(ctx context.Context, fn func(interfaces.KanbanRepository) error) error {
    // begin a tx on the underlying pool
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }

    // create a repo backed by the tx executor
    txRepo := &KanbanRepo{
        exec: TxExecutor{Tx: tx},
        pool: r.pool, // keep pool in case nested Tx needs it (we won't start nested tx)
    }

    // call the user function
    if err := fn(txRepo); err != nil {
        _ = tx.Rollback(ctx)
        return err
    }

    return tx.Commit(ctx)
}
