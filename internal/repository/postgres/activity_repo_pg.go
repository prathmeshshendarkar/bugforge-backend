package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepoPG struct {
	db *pgxpool.Pool
}

func NewActivityRepository(db *pgxpool.Pool) repo.ActivityRepository {
	return &ActivityRepoPG{db: db}
}

func (r *ActivityRepoPG) Create(ctx context.Context, a *models.IssueActivity) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}

	var meta json.RawMessage
	if a.Metadata != nil {
		b, err := json.Marshal(a.Metadata)
		if err != nil {
			return err
		}
		meta = b
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO issue_activity_logs (id, issue_id, user_id, action, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`,
		a.ID, a.IssueID, a.UserID, a.Action, meta)

	return err
}

func (r *ActivityRepoPG) ListByIssue(ctx context.Context, issueID string) ([]models.IssueActivity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, issue_id, user_id, action, metadata, created_at
		 FROM issue_activity_logs
		 WHERE issue_id = $1
		 ORDER BY created_at ASC`,
		issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueActivity

	for rows.Next() {
		var a models.IssueActivity
		var meta json.RawMessage

		if err := rows.Scan(
			&a.ID,
			&a.IssueID,
			&a.UserID,
			&a.Action,
			&meta,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}

		if len(meta) > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal(meta, &m); err == nil {
				a.Metadata = m
			} else {
				a.Metadata = map[string]interface{}{}
			}
		}

		out = append(out, a)
	}

	return out, nil
}

func (r *ActivityRepoPG) ListByProject(ctx context.Context, projectID string) ([]models.IssueActivity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT 
			a.id, a.issue_id, a.user_id, a.action, a.metadata, a.created_at,
			i.title AS issue_title
		FROM issue_activity_logs a
		LEFT JOIN issues i ON a.issue_id = i.id
		WHERE i.project_id = $1
		ORDER BY a.created_at DESC;
		`,
		projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueActivity

	for rows.Next() {
		var a models.IssueActivity
		var meta json.RawMessage

		var issueTitle *string

		if err := rows.Scan(
			&a.ID,
			&a.IssueID,
			&a.UserID,
			&a.Action,
			&meta,
			&a.CreatedAt,
			&issueTitle,
		); err != nil {
			return nil, err
		}

		a.IssueTitle = issueTitle

		if len(meta) > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal(meta, &m); err == nil {
				a.Metadata = m
			} else {
				a.Metadata = map[string]interface{}{}
			}
		}

		out = append(out, a)
	}

	return out, nil
}
