package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepoPG struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) repo.UserRepository {
	return &UserRepoPG{db: db}
}

func (r *UserRepoPG) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (id, customer_id, name, username, email, password_hash, role, default_project_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.CustomerID, u.Name, u.Username, u.Email, u.PasswordHash, u.Role, u.DefaultProjectID,
	)
	return err
}

func (r *UserRepoPG) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, customer_id, name, username, email, password_hash, role, default_project_id, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.CustomerID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.DefaultProjectID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepoPG) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, customer_id, name, username, email, password_hash, role, default_project_id, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.CustomerID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.DefaultProjectID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepoPG) GetByUsername(ctx context.Context, username string) (*models.User, error) {
    row := r.db.QueryRow(ctx,
        `SELECT id, customer_id, email, username, created_at 
         FROM users
         WHERE username = $1`,
        username,
    )

    var u models.User
    if err := row.Scan(&u.ID, &u.CustomerID, &u.Email, &u.Username, &u.CreatedAt); err != nil {
        return nil, err
    }

    return &u, nil
}

func (r *UserRepoPG) GetAllByCustomer(ctx context.Context, customerID string) ([]models.User, error) {
	query := `
		SELECT id, customer_id, name, username, email, password_hash, role, default_project_id, created_at, updated_at
		FROM users
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID,
			&u.CustomerID,
			&u.Name,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.Role,
			&u.DefaultProjectID,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepoPG) Update(ctx context.Context, u *models.User) error {
	query := `
		UPDATE users
		SET name = $1, username = $2, email = $3, password_hash = $4, role = $5,
			default_project_id = $6, updated_at = $7
		WHERE id = $8 AND customer_id = $9
	`
	_, err := r.db.Exec(ctx, query, u.Name, u.Username, u.Email, u.PasswordHash, u.Role, u.DefaultProjectID, time.Now(), u.ID, u.CustomerID)
	return err
}

func (r *UserRepoPG) Delete(ctx context.Context, id, customerID string) error {
	// hard delete; if you want soft delete, add an "is_active" or "deleted_at" column to users
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// remove assignments
	_, err = tx.Exec(ctx, `DELETE FROM project_members WHERE user_id = $1`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM users WHERE id = $1 AND customer_id = $2`, id, customerID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *UserRepoPG) AssignProjects(ctx context.Context, userID string, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// delete existing assignments, then insert new ones
	_, err = tx.Exec(ctx, `DELETE FROM project_members WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO project_members (user_id, project_id) VALUES ($1, $2)`
	for _, pid := range projectIDs {
		if _, err := tx.Exec(ctx, stmt, userID, pid); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepoPG) DeleteProjectAssignments(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_members WHERE user_id = $1`, userID)
	return err
}

func (r *UserRepoPG) GetAssignedProjectIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT project_id FROM project_members WHERE user_id = $1 ORDER BY project_id`, userID)
	fmt.Println(rows);
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		ids = append(ids, pid)
	}
	return ids, nil
}

func (r *UserRepoPG) CreatePending(ctx context.Context, u *models.User) error {
    query := `
        INSERT INTO users (
            id,
            customer_id,
            name,
            username,
            email,
            role,
            is_pending,
            created_at,
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
    `
    _, err := r.db.Exec(ctx, query,
        u.ID,
        u.CustomerID,
        u.Name,
        u.Username,
        u.Email,
        u.Role,
    )
    return err
}


func (r *UserRepoPG) SaveInviteToken(ctx context.Context, userID, token string) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO user_invites (user_id, token, created_at)
         VALUES ($1, $2, NOW())
         ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token`,
        userID, token,
    )
    return err
}

func (r *UserRepoPG) GetByInviteToken(ctx context.Context, token string) (*models.User, error) {
    query := `
        SELECT u.id, u.customer_id, u.name, u.email, u.password_hash, 
               u.role, u.default_project_id, u.is_pending, u.created_at, u.updated_at
        FROM user_invites ui
        JOIN users u ON ui.user_id = u.id
        WHERE ui.token = $1 AND ui.expires_at > NOW()
    `
    var u models.User

    err := r.db.QueryRow(ctx, query, token).Scan(
        &u.ID, &u.CustomerID, &u.Name, &u.Email,
        &u.PasswordHash, &u.Role, &u.DefaultProjectID,
        &u.IsPending, &u.CreatedAt, &u.UpdatedAt,
    )

    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    return &u, nil
}

func (r *UserRepoPG) MarkInviteAccepted(ctx context.Context, userID string) error {
    _, err := r.db.Exec(ctx,
        `DELETE FROM user_invites WHERE user_id = $1`,
        userID,
    )
    return err
}

func (r *UserRepoPG) DeleteInvitesByUserID(ctx context.Context, userID string) error {
    _, err := r.db.Exec(ctx, `
        DELETE FROM user_invites
        WHERE user_id = $1
    `, userID)
    return err
}