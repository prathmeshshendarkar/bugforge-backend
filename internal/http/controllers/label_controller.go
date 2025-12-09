package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type ProjectLabelControllerImpl struct {
	svc service.LabelService
}

func NewProjectLabelController(s service.LabelService) interfaces.ProjectLabelController {
	return &ProjectLabelControllerImpl{svc: s}
}

type createLabelReq struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (pc *ProjectLabelControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	projectID := c.Params("project_id")
	var req createLabelReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	l, err := pc.svc.CreateLabel(context.Background(), customerID.(string), projectID, req.Name, req.Color, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, l)
}

func (pc *ProjectLabelControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	projectID := c.Params("project_id")

	out, err := pc.svc.ListLabelsByProject(context.Background(), customerID.(string), projectID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

type updateLabelReq struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

func (pc *ProjectLabelControllerImpl) Update(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	projectID := c.Params("project_id")
	labelID := c.Params("label_id")

	var req updateLabelReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}
	var name, color string
	if req.Name != nil {
		name = *req.Name
	}
	if req.Color != nil {
		color = *req.Color
	}

	l, err := pc.svc.UpdateLabel(context.Background(), customerID.(string), projectID, labelID, name, color, userID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, l)
}

func (pc *ProjectLabelControllerImpl) Delete(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	projectID := c.Params("project_id")
	labelID := c.Params("label_id")

	if err := pc.svc.DeleteLabel(context.Background(), customerID.(string), projectID, labelID, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"deleted": true})
}
