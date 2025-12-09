package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) repo.ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}

func (r *ProjectRepositoryImpl) Create(ctx context.Context, p *models.Project) error {
	query := `
        INSERT INTO projects (id, customer_id, name, slug, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `
	_, err := r.db.Exec(ctx, query, p.ID, p.CustomerID, p.Name, p.Slug)
	return err
}

func (r *ProjectRepositoryImpl) GetAll(ctx context.Context, customerID string) ([]models.Project, error) {
	query := `
        SELECT id, customer_id, name, slug, created_at, updated_at
        FROM projects
        WHERE customer_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project

	for rows.Next() {
		var p models.Project
		if err := rows.Scan(
			&p.ID,
			&p.CustomerID,
			&p.Name,
			&p.Slug,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *ProjectRepositoryImpl) GetByID(ctx context.Context, id string, customerID string) (*models.Project, error) {
	query := `
        SELECT id, customer_id, name, slug, created_at, updated_at
        FROM projects
        WHERE id = $1 AND customer_id = $2
        LIMIT 1
    `

	var p models.Project
	err := r.db.QueryRow(ctx, query, id, customerID).Scan(
		&p.ID,
		&p.CustomerID,
		&p.Name,
		&p.Slug,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}


func (r *ProjectRepositoryImpl) GetBySlug(ctx context.Context, slug string, customerID string) (*models.Project, error) {
	query := `
        SELECT id, customer_id, name, slug, created_at, updated_at
        FROM projects
        WHERE slug = $1 AND customer_id = $2
        LIMIT 1
    `

	var p models.Project
	err := r.db.QueryRow(ctx, query, slug, customerID).Scan(
		&p.ID,
		&p.CustomerID,
		&p.Name,
		&p.Slug,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}


func (r *ProjectRepositoryImpl) Update(ctx context.Context, p *models.Project) error {
	query := `
        UPDATE projects
        SET name = $1, slug = $2, updated_at = NOW()
        WHERE id = $3 AND customer_id = $4
    `
	_, err := r.db.Exec(ctx, query, p.Name, p.Slug, p.ID, p.CustomerID)
	return err
}


func (r *ProjectRepositoryImpl) Delete(ctx context.Context, id string, customerID string) error {
	query := `DELETE FROM projects WHERE id = $1 AND customer_id = $2`
	_, err := r.db.Exec(ctx, query, id, customerID)
	return err
}
