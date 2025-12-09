package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type IssueCommentControllerImpl struct {
	svc service.IssueService
}

func NewIssueCommentController(s service.IssueService) interfaces.IssueCommentController {
	return &IssueCommentControllerImpl{svc: s}
}

type createCommentReq struct {
	Body string `json:"body"`
}

func (ic *IssueCommentControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	issueID := c.Params("id")

	var req createCommentReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	comment, err := ic.svc.CreateComment(context.Background(), customerID.(string), issueID, userID.(string), req.Body)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, comment)
}

func (ic *IssueCommentControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")
	out, err := ic.svc.ListComments(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}

type updateCommentReq struct {
	Body string `json:"body"`
}

func (ic *IssueCommentControllerImpl) Update(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	commentID := c.Params("comment_id")
	var req updateCommentReq
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "invalid payload")
	}
	comment, err := ic.svc.UpdateComment(context.Background(), customerID.(string), commentID, userID.(string), req.Body)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, comment)
}

func (ic *IssueCommentControllerImpl) Delete(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	commentID := c.Params("comment_id")
	if err := ic.svc.DeleteComment(context.Background(), customerID.(string), commentID, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"deleted": true})
}
