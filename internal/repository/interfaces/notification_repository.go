package interfaces

import "bugforge-backend/internal/models"

type NotificationRepository interface {
    Create(notification *models.Notification) error
    GetByUser(userID string) ([]models.Notification, error)
    MarkRead(id string) error
    MarkAllRead(userID string) error
}
