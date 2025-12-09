package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

type IssueSubtaskControllerImpl struct {
	svc service.IssueService
}

func NewIssueSubtaskController(s service.IssueService) interfaces.IssueSubtaskController {
	return &IssueSubtaskControllerImpl{svc: s}
}

type createSubtaskReq struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	AssignedTo  *string `json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
}

func (is *IssueSubtaskControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	issueID := c.Params("id")
	var req createSubtaskReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	sub, err := is.svc.CreateSubtask(context.Background(), customerID.(string), issueID, req.Title, req.Description, req.AssignedTo, req.DueDate, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, sub)
}

type updateSubtaskReq struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Status      *string   `json:"status"`
	AssignedTo  *string   `json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
}

func (is *IssueSubtaskControllerImpl) Update(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	subtaskID := c.Params("subtask_id")
	var req updateSubtaskReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	// map values
	var title, status string
	var description, assignedTo *string
	var dueDate *time.Time

	if req.Title != nil {
		title = *req.Title
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.Description != nil {
		description = req.Description
	}
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	sub, err := is.svc.UpdateSubtask(context.Background(), customerID.(string), subtaskID, title, description, status, assignedTo, dueDate, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, sub)
}

func (is *IssueSubtaskControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	out, err := is.svc.ListSubtasks(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

func (is *IssueSubtaskControllerImpl) Delete(c *fiber.Ctx) error {
    customerID := c.Locals("customer_id")
    if customerID == nil {
        return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
    }

    subtaskID := c.Params("subtask_id")

    err := is.svc.DeleteSubtask(context.Background(), customerID.(string), subtaskID)
    if err != nil {
        return helpers.Error(c, fiber.StatusBadRequest, err.Error())
    }

    return helpers.Success(c, fiber.Map{"deleted": true})
}
