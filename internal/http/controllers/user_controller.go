package controllers

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type UserControllerImpl struct {
	userService service.UserService
}

func NewUserController(us service.UserService) controller.UserController {
	return &UserControllerImpl{userService: us}
}

// DTOs
type CreateUserRequest struct {
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	Password          string   `json:"password"`
	Role              string   `json:"role"`
	AssignedProjectIDs []string `json:"assigned_project_ids"`
	DefaultProjectID  *string  `json:"default_project_id"`
}

type UpdateUserRequest struct {
	Name              *string  `json:"name"`
	Email             *string  `json:"email"`
	Password          *string  `json:"password"`
	Role              *string  `json:"role"`
	AssignedProjectIDs []string `json:"assigned_project_ids"`
	DefaultProjectID  *string  `json:"default_project_id"`
}

// Create user (super_admin scoped)
func (uc *UserControllerImpl) CreateUser(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var body CreateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request")
	}

	u, err := uc.userService.CreateUser(context.Background(), customerID.(string), body.Name, body.Email, body.Password, body.Role, body.AssignedProjectIDs, body.DefaultProjectID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	// hide password hash from response
	u.PasswordHash = helpers.StrPtr("")
	return helpers.Success(c, u)
}

func (uc *UserControllerImpl) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	u, err := uc.userService.GetByID(context.Background(), id, customerID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusNotFound, err.Error())
	}
	if u == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	u.PasswordHash = helpers.StrPtr("")
	return helpers.Success(c, u)
}

func (uc *UserControllerImpl) GetAllUsers(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	users, err := uc.userService.GetAllByCustomer(context.Background(), customerID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	// clear password hashes
	for i := range users {
		users[i].PasswordHash = helpers.StrPtr("")
	}
	return helpers.Success(c, users)
}

func (uc *UserControllerImpl) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var body UpdateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request")
	}

	// pull values safely
	var name, email, password, role string
	if body.Name != nil {
		name = *body.Name
	}
	if body.Email != nil {
		email = *body.Email
	}
	if body.Password != nil {
		password = *body.Password
	}
	if body.Role != nil {
		role = *body.Role
	}

	u, err := uc.userService.UpdateUser(context.Background(), id, customerID.(string), name, email, password, role, body.AssignedProjectIDs, body.DefaultProjectID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	u.PasswordHash = helpers.StrPtr("")
	return helpers.Success(c, u)
}

func (uc *UserControllerImpl) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	if err := uc.userService.DeleteUser(context.Background(), id, customerID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return helpers.Success(c, fiber.Map{"deleted": true})
}
