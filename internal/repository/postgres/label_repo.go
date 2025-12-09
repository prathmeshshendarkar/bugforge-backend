package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LabelRepoPG struct {
	db *pgxpool.Pool
}

func NewLabelRepository(db *pgxpool.Pool) repo.LabelRepository {
	return &LabelRepoPG{db: db}
}


func (r *LabelRepoPG) CreateLabel(ctx context.Context, l *models.Label) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO labels (id, customer_id, project_id, name, color, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, l.ID, l.CustomerID, l.ProjectID, l.Name, l.Color)
	return err
}


func (r *LabelRepoPG) UpdateLabel(ctx context.Context, l *models.Label) error {
	_, err := r.db.Exec(ctx, `
		UPDATE labels 
		SET name = $1, color = $2 
		WHERE id = $3
	`, l.Name, l.Color, l.ID)
	return err
}


func (r *LabelRepoPG) DeleteLabel(ctx context.Context, labelID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM labels WHERE id=$1`, labelID)
	return err
}

func (r *LabelRepoPG) GetLabelByID(ctx context.Context, id string) (*models.Label, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, customer_id, project_id, name, color, created_at
		FROM labels WHERE id = $1
	`, id)

	var l models.Label
	err := row.Scan(&l.ID, &l.CustomerID, &l.ProjectID, &l.Name, &l.Color, &l.CreatedAt)
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &l, err
}


func (r *LabelRepoPG) ListLabelsByProject(ctx context.Context, projectID string) ([]models.Label, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, project_id, name, color, created_at
		FROM labels WHERE project_id = $1
		ORDER BY created_at ASC
	`, projectID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Label

	for rows.Next() {
		var l models.Label
		err := rows.Scan(&l.ID, &l.CustomerID, &l.ProjectID, &l.Name, &l.Color, &l.CreatedAt)

		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}

	return out, nil
}
