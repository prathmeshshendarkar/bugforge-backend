package controllers

import (
	ctrlIface "bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	iface "bugforge-backend/internal/service/interfaces"

	"github.com/gofiber/fiber/v2"
)

type NotificationControllerImpl struct {
	ns iface.NotificationService
}

func NewNotificationController(ns iface.NotificationService) ctrlIface.NotificationController {
	return &NotificationControllerImpl{ns: ns}
}

func (ctrl *NotificationControllerImpl) GetNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "user_id is required")
	}

	notifications, err := ctrl.ns.GetUserNotifications(userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "failed to fetch notifications")
	}

	return c.JSON(fiber.Map{"data": notifications})
}

func (ctrl *NotificationControllerImpl) MarkRead(c *fiber.Ctx) error {
	nID := c.Params("id")
	if nID == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "id is required")
	}

	if err := ctrl.ns.MarkAsRead(nID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "failed to mark as read")
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func (ctrl *NotificationControllerImpl) MarkAllRead(c *fiber.Ctx) error {
    userID := c.Query("user_id")
    if userID == "" {
        return helpers.Error(c, fiber.StatusBadRequest, "user_id is required")
    }

    if err := ctrl.ns.MarkAllAsRead(userID); err != nil {
        return helpers.Error(c, fiber.StatusInternalServerError, "failed to mark all as read")
    }

    return c.JSON(fiber.Map{"status": "ok"})
}
