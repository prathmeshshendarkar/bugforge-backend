package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type IssueChecklistControllerImpl struct {
	svc service.IssueService
}

func NewIssueChecklistController(s service.IssueService) interfaces.IssueChecklistController {
	return &IssueChecklistControllerImpl{svc: s}
}

type createChecklistReq struct {
	Title string `json:"title"`
}

func (ic *IssueChecklistControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")

	var req createChecklistReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	cl, err := ic.svc.CreateChecklist(context.Background(), customerID.(string), issueID, req.Title, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, cl)
}

type addChecklistItemReq struct {
	Content string `json:"content"`
}

func (ic *IssueChecklistControllerImpl) AddItem(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	checklistID := c.Params("checklist_id")
	var req addChecklistItemReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	item, err := ic.svc.CreateChecklistItem(context.Background(), customerID.(string), checklistID, req.Content, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, item)
}

type updateChecklistItemReq struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

func (ic *IssueChecklistControllerImpl) UpdateItem(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	itemID := c.Params("item_id")
	var req updateChecklistItemReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	item, err := ic.svc.UpdateChecklistItem(context.Background(), customerID.(string), itemID, req.Content, req.Done, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, item)
}

func (ic *IssueChecklistControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	out, err := ic.svc.ListChecklists(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

func (ic *IssueChecklistControllerImpl) DeleteChecklist(c *fiber.Ctx) error {
	checklistID := c.Params("checklist_id")
	customerID := c.Locals("customer_id")

	if customerID == nil {
		return helpers.Error(c, 401, "unauthorized")
	}

	if err := ic.svc.DeleteChecklist(context.Background(), customerID.(string), checklistID); err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, fiber.Map{"deleted": true})
}

func (ic *IssueChecklistControllerImpl) DeleteItem(c *fiber.Ctx) error {
	itemID := c.Params("item_id")
	customerID := c.Locals("customer_id")

	if customerID == nil {
		return helpers.Error(c, 401, "unauthorized")
	}

	if err := ic.svc.DeleteChecklistItem(context.Background(), customerID.(string), itemID); err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, fiber.Map{"deleted": true})
}

type reorderReq struct {
	Order []struct {
		ItemID     string `json:"itemId"`
		OrderIndex int    `json:"orderIndex"`
	} `json:"order"`
}

func (ic *IssueChecklistControllerImpl) ReorderItems(c *fiber.Ctx) error {
	var req reorderReq
	customerID := c.Locals("customer_id")
	checklistID := c.Params("checklist_id")

	if customerID == nil {
		return helpers.Error(c, 401, "unauthorized")
	}

	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, 400, "invalid payload")
	}

	items := make([]models.ChecklistItem, len(req.Order))
	for i, r := range req.Order {
		items[i] = models.ChecklistItem{
			ID:         r.ItemID,
			OrderIndex: r.OrderIndex,
		}
	}

	err := ic.svc.ReorderChecklistItems(context.Background(), customerID.(string), checklistID, items)
	if err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, fiber.Map{"reordered": true})
}

