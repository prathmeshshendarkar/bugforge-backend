package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
)

type IssueControllerImpl struct {
	svc service.IssueService
}

func NewIssueController(s service.IssueService) interfaces.IssueController {
	return &IssueControllerImpl{svc: s}
}

type createIssueReq struct {
	ProjectID   string  `json:"project_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	AssignedTo  *string `json:"assigned_to"`
}

// @Summary Create a new issue
// @Tags Issues
// @Param data body createIssueReq true "Issue data"
// @Success 200 {object} map[string]interface{}
// @Router /issues [post]
func (it *IssueControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var req createIssueReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	issue, err := it.svc.CreateIssue(context.Background(), customerID.(string), req.ProjectID, req.Title, req.Description, req.Priority, req.AssignedTo, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, issue)
}

func (it *IssueControllerImpl) Get(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	id := c.Params("id")
	issue, err := it.svc.GetIssue(context.Background(), customerID.(string), id)
	if err != nil {
		return helpers.Error(c, fiber.StatusNotFound, err.Error())
	}
	return helpers.Success(c, issue)
}

func (it *IssueControllerImpl) ListAll(c *fiber.Ctx) error {
    customerID := c.Locals("customer_id")
    if customerID == nil {
        return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
    }

    issues, err := it.svc.ListAllIssues(context.Background(), customerID.(string))
    if err != nil {
        return helpers.Error(c, fiber.StatusBadRequest, err.Error())
    }

    return helpers.Success(c, issues)
}



func (it *IssueControllerImpl) ListByProject(c *fiber.Ctx) error {
    projectID := c.Params("project_id")
    customerID := c.Locals("customer_id").(string)

    queries := c.Queries() // map[string]string

	values := url.Values{}
	for k, v := range queries {
		values.Set(k, v)
	}

	issues, err := it.svc.ListIssuesByProject(
		c.Context(),
		projectID,
		customerID,
		values,
	)

    if err != nil {
        return helpers.Error(c, fiber.StatusBadRequest, err.Error())
    }

    return helpers.Success(c, issues)
}


type updateIssueReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssignedTo  *string `json:"assigned_to"`

	AssignedToSnake *string `json:"assigned_to"`
    AssignedToCamel *string `json:"assignedTo"`
}

func (it *IssueControllerImpl) Update(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	id := c.Params("id")
	var req updateIssueReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	var title, description, status, priority string
	var assignedTo *string

	if req.AssignedToSnake != nil {
		assignedTo = req.AssignedToSnake
	}
	if req.AssignedToCamel != nil {
		assignedTo = req.AssignedToCamel
	}
	fmt.Println("Updated issues ", assignedTo);
	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.Priority != nil {
		priority = *req.Priority
	}
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}

	issue, err := it.svc.UpdateIssue(context.Background(), customerID.(string), id, title, description, status, priority, assignedTo, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, issue)
}

func (it *IssueControllerImpl) Delete(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	id := c.Params("id")
	if err := it.svc.DeleteIssue(context.Background(), customerID.(string), id, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"deleted": true})
}


func (it *IssueControllerImpl) ListActivity(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	out, err := it.svc.ListActivity(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

func (c *IssueControllerImpl) UpdateDueDate(ctx *fiber.Ctx) error {
	issueID := ctx.Params("id")
	customerID := ctx.Locals("customer_id").(string)
	userID := ctx.Locals("user_id").(string)

	var body struct {
		DueDate *time.Time `json:"due_date"`
	}

	if err := ctx.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	if err := c.svc.UpdateDueDate(ctx.Context(), customerID, issueID, body.DueDate, userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"due_date": body.DueDate,
	})
}
