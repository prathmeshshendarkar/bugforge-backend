package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectMemberRepoPG struct {
    db *pgxpool.Pool
}

func NewProjectMemberRepo(db *pgxpool.Pool) repo.ProjectMemberRepository {
    return &ProjectMemberRepoPG{db: db}
}

func (r *ProjectMemberRepoPG) AddMember(ctx context.Context, projectID, userID string) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO project_members (project_id, user_id)
         VALUES ($1, $2) ON CONFLICT DO NOTHING`,
        projectID, userID,
    )
    return err
}

func (r *ProjectMemberRepoPG) RemoveMember(ctx context.Context, projectID, userID string) error {
    _, err := r.db.Exec(ctx,
        `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`,
        projectID, userID,
    )
    return err
}

func (r *ProjectMemberRepoPG) ListMembers(ctx context.Context, projectID, customerID string) ([]models.User, error) {
    rows, err := r.db.Query(ctx,
        `SELECT u.id, u.name, u.email, u.created_at, u.updated_at
         FROM project_members pm
         JOIN users u ON pm.user_id = u.id
         WHERE pm.project_id = $1 AND u.customer_id = $2`,
        projectID, customerID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    members := []models.User{}
    for rows.Next() {
        var u models.User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        members = append(members, u)
    }

    return members, nil
}

func (r *ProjectMemberRepoPG) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
    var exists bool
    err := r.db.QueryRow(ctx,
        `SELECT EXISTS(
            SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2
        )`,
        projectID, userID,
    ).Scan(&exists)
    return exists, err
}

// GetAssignedProjectIDsForUser returns project IDs the given user is a member of
func (r *ProjectMemberRepoPG) GetAssignedProjectIDsForUser(ctx context.Context, userID string) ([]string, error) {
    rows, err := r.db.Query(ctx,
        `SELECT project_id FROM project_members WHERE user_id = $1`,
        userID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := []string{}
    for rows.Next() {
        var pid string
        if err := rows.Scan(&pid); err != nil {
            return nil, err
        }
        out = append(out, pid)
    }
    return out, nil
}

// SyncMembersForUser replaces the user's memberships for a customer with the given list.
// It removes existing memberships (for that customer) and inserts the new ones in a transaction.
func (r *ProjectMemberRepoPG) SyncMembersForUser(ctx context.Context, userID, customerID string, projectIDs []string) error {
    tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return err
    }

    defer func() {
        if err != nil {
            _ = tx.Rollback(ctx)
        }
    }()

    // Remove all memberships for this user under this customer's projects
    _, err = tx.Exec(ctx, `
        DELETE FROM project_members 
        WHERE user_id = $1
        AND project_id IN (
            SELECT id FROM projects WHERE customer_id = $2
        )
    `, userID, customerID)
    if err != nil {
        return err
    }

    // Insert new memberships
    for _, pid := range projectIDs {
        _, err = tx.Exec(ctx, `
            INSERT INTO project_members (project_id, user_id)
            VALUES ($1, $2)
            ON CONFLICT DO NOTHING
        `, pid, userID)
        if err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}
