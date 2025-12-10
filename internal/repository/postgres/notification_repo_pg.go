package postgres

import (
	"context"
	"fmt"
	"time"

	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepoPG struct {
	pool *pgxpool.Pool
}

func NewNotificationRepoPG(pool *pgxpool.Pool) repo.NotificationRepository {
	return &NotificationRepoPG{pool: pool}
}

func (r *NotificationRepoPG) Create(notification *models.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	INSERT INTO notifications (id, user_id, type, title, message, metadata, is_read, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Metadata,
		notification.IsRead,
		notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("notification create: %w", err)
	}
	return nil
}

func (r *NotificationRepoPG) GetByUser(userID string) ([]models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT id, user_id, type, title, message, metadata, is_read, created_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.Metadata,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *NotificationRepoPG) MarkRead(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE notifications SET is_read = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *NotificationRepoPG) MarkAllRead(userID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `UPDATE notifications SET is_read = true WHERE user_id = $1`
    _, err := r.pool.Exec(ctx, query, userID)
    return err
}
