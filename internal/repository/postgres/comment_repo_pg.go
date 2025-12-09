package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepoPG struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) repo.CommentRepository {
	return &CommentRepoPG{db: db}
}

func (r *CommentRepoPG) Create(ctx context.Context, c *models.IssueComment) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	_, err := r.db.Exec(ctx, `INSERT INTO issue_comments (id, issue_id, user_id, body, created_at) VALUES ($1,$2,$3,$4,NOW())`,
		c.ID, c.IssueID, c.UserID, c.Body)
	return err
}

func (r *CommentRepoPG) ListByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error) {
	rows, err := r.db.Query(ctx, `SELECT id, issue_id, user_id, body, created_at FROM issue_comments WHERE issue_id = $1 ORDER BY created_at ASC`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueComment
	for rows.Next() {
		var c models.IssueComment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
