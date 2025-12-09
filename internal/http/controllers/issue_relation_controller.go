package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type IssueRelationControllerImpl struct {
	svc service.IssueService
}

func NewIssueRelationController(s service.IssueService) interfaces.IssueRelationController {
	return &IssueRelationControllerImpl{svc: s}
}

type addRelationReq struct {
	RelatedIssueID string `json:"related_issue_id"`
	RelationType   string `json:"relation_type"`
}

func (ir *IssueRelationControllerImpl) Add(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	var req addRelationReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}
	if err := ir.svc.AddRelation(context.Background(), customerID.(string), issueID, req.RelatedIssueID, req.RelationType, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"added": true})
}

func (ir *IssueRelationControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	out, err := ir.svc.ListRelations(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

func (ir *IssueRelationControllerImpl) Delete(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	relationID := c.Params("relation_id")
	if err := ir.svc.DeleteRelation(context.Background(), customerID.(string), relationID, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"deleted": true})
}
