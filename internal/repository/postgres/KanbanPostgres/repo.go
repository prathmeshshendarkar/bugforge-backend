package postgres

import (
	"bugforge-backend/internal/repository/interfaces"

	"github.com/jackc/pgx/v5/pgxpool"
)

// KanbanRepo implements interfaces.KanbanRepository using a DBExecutor.
type KanbanRepo struct {
    exec DBExecutor      // the executor used for queries (pool or tx)
    pool *pgxpool.Pool   // pool is kept so we can start transactions when needed
}

// NewKanbanRepo constructs a non-transactional repo using the pool.
func NewKanbanRepo(pool *pgxpool.Pool) interfaces.KanbanRepository {
    return &KanbanRepo{
        exec: PoolExecutor{DB: pool},
        pool: pool,
    }
}
