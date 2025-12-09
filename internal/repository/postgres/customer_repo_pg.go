package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepoPG struct {
    db *pgxpool.Pool
}

func NewCustomerRepo(db *pgxpool.Pool) repo.CustomerRepository {
    return &CustomerRepoPG{db}
}

func HealthCheck() string{
	return "Working Fine"
}

func (r *CustomerRepoPG) GetByID(ctx context.Context, id string) (*models.Customer, error) {
    query := `
        SELECT id, name, created_at, updated_at
        FROM customers
        WHERE id = $1;
    `
    row := r.db.QueryRow(ctx, query, id)

    var c models.Customer
    err := row.Scan(&c.Id, &c.Name, &c.CreatedAt, &c.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &c, nil
}
