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
	_, err := r.db.Exec(ctx,
		`INSERT INTO issue_comments (id, issue_id, user_id, body, body_html, created_at)
		VALUES ($1,$2,$3,$4,$5,NOW())`,
		c.ID, c.IssueID, c.UserID, c.Body, c.BodyHTML,
	)

	return err
}

func (r *CommentRepoPG) ListCommentsByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT 
			c.id,
			c.issue_id,
			c.user_id,
			c.body,
			c.body_html,
			c.created_at,
			c.updated_at,
			u.name AS author_name,
			u.email AS author_email
		FROM issue_comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.issue_id = $1
		ORDER BY c.created_at ASC;`,
		issueID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueComment
	for rows.Next() {
		var c models.IssueComment
		if err := rows.Scan(
			&c.ID,
			&c.IssueID,
			&c.UserID,
			&c.Body,
			&c.BodyHTML,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.AuthorName,
			&c.AuthorEmail,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *CommentRepoPG) Update(ctx context.Context, c *models.IssueComment) error {
    _, err := r.db.Exec(ctx,
        `UPDATE issue_comments
         SET body = $1, body_html = $2, updated_at = NOW()
         WHERE id = $3`,
        c.Body, c.BodyHTML, c.ID,
    )
    return err
}
