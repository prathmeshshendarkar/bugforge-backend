package service

import (
	"context"
	"fmt"
	"time"

	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	iface "bugforge-backend/internal/service/interfaces"
	"bugforge-backend/internal/websocket/notifications"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NotificationServiceImpl struct {
	repo     repo.NotificationRepository
	userRepo repo.UserRepository
	wsHub *notifications.NotificationHub
}

func NewNotificationService(
	repo repo.NotificationRepository,
	userRepo repo.UserRepository,
	wsHub *notifications.NotificationHub,
) iface.NotificationService {
	return &NotificationServiceImpl{
		repo:     repo,
		userRepo: userRepo,
		wsHub:    wsHub,
	}
}

func (s *NotificationServiceImpl) SendInApp(userID, title, message, metadata string) error {
	n := &models.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      "in-app",
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(n); err != nil {
		return fmt.Errorf("send in-app: %w", err)
	}

	// Send through WebSocket if available
	if s.wsHub != nil {
		s.wsHub.Broadcast <- notifications.NotificationWSMessage{
			UserID: userID,
			Type:   "in-app",
			Data: fiber.Map{
				"title":    title,
				"message":  message,
				"metadata": metadata,
			},
		}
	}

	return nil
}

func (s *NotificationServiceImpl) SendEmail(userID, title, message string) error {
	ctx := context.Background()

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}

	if user.Email == "" {
		return fmt.Errorf("user has no email")
	}

	if err := helpers.SendEmail(user.Email, title, message); err != nil {
		return fmt.Errorf("email send failed: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) MarkAsRead(notificationID string) error {
	return s.repo.MarkRead(notificationID)
}

func (s *NotificationServiceImpl) GetUserNotifications(userID string) ([]iface.NotificationView, error) {
	ns, err := s.repo.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	out := make([]iface.NotificationView, 0, len(ns))
	for _, n := range ns {
		out = append(out, iface.NotificationView{
			ID:        n.ID,
			UserID:    n.UserID,
			Type:      n.Type,
			Title:     n.Title,
			Message:   n.Message,
			Metadata:  n.Metadata,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		})
	}

	return out, nil
}

func (s *NotificationServiceImpl) MarkAllAsRead(userID string) error {
    return s.repo.MarkAllRead(userID)
}
