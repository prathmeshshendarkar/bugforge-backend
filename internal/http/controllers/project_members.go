package controllers

import (
	"bugforge-backend/internal/http/helpers"
	svc "bugforge-backend/internal/service/interfaces"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ProjectMemberController struct {
	service svc.ProjectMemberService
}

func NewProjectMemberController(s svc.ProjectMemberService) *ProjectMemberController {
	return &ProjectMemberController{service: s}
}

func (pc *ProjectMemberController) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id").(string)
	projectID := c.Params("project_id")

	out, err := pc.service.ListMembers(c.Context(), projectID, customerID)
	if err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, out)
}

type addMemberReq struct {
	UserID string `json:"userId"`
}

type inviteReq struct {
    Email string `json:"email"`
    Role  string `json:"role"`
}

func (pc *ProjectMemberController) Add(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id").(string)
	projectID := c.Params("project_id")

	var body addMemberReq
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, 400, "invalid request")
	}

	fmt.Println(body);
	fmt.Println(customerID);
	err := pc.service.AddMember(c.Context(), projectID, customerID, body.UserID)
	if err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, fiber.Map{"added": true})
}

func (pc *ProjectMemberController) Remove(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id").(string)
	projectID := c.Params("project_id")
	userID := c.Params("user_id")

	_ = customerID // Not needed, but preserved for consistency

	err := pc.service.RemoveMember(c.Context(), projectID, customerID, userID)
	if err != nil {
		return helpers.Error(c, 400, err.Error())
	}

	return helpers.Success(c, fiber.Map{"removed": true})
}

func (pc *ProjectMemberController) Invite(c *fiber.Ctx) error {
    customerID := c.Locals("customer_id").(string)
    projectID := c.Params("project_id")

    var body inviteReq
    if err := c.BodyParser(&body); err != nil {
        return helpers.Error(c, 400, "invalid request body")
    }

    err := pc.service.Invite(c.Context(), projectID, customerID, body.Email, body.Role)
    if err != nil {
        return helpers.Error(c, 400, err.Error())
    }

    return helpers.Success(c, fiber.Map{"invited": true})
}