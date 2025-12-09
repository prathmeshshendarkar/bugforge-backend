package controllers

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AuthControllerImpl struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthController(as service.AuthService, us service.UserService) controller.AuthController {
	return &AuthControllerImpl{
		authService: as,
		userService: us,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary Login
// @Tags Auth
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (ac *AuthControllerImpl) Login(c *fiber.Ctx) error {
	var body LoginRequest

	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid payload")
	}

	user, token, err := ac.authService.Login(context.Background(), body.Email, body.Password)
	if err != nil {
		return helpers.Error(c, fiber.StatusUnauthorized, err.Error())
	}

	user.PasswordHash = nil

	return helpers.Success(c, fiber.Map{
		"user":  user,
		"token": token,
	})
}

func (ac *AuthControllerImpl) Me(c *fiber.Ctx) error {
    userID := c.Locals("user_id")
    customerID := c.Locals("customer_id")

    if userID == nil || customerID == nil {
		fmt.Println("Error ", userID, customerID);
        return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
    }

    // Fetch user from DB
    user, err := ac.userService.GetByID(
        c.Context(),
        userID.(string),
        customerID.(string),
    )
	fmt.Println(user);
    if err != nil {
		fmt.Println("Error ", err);
        return helpers.Error(c, fiber.StatusInternalServerError, err.Error())
    }
    if user == nil {
		fmt.Println(user);
        return helpers.Error(c, fiber.StatusNotFound, "User not found")
    }

    user.PasswordHash = nil

    return helpers.Success(c, user)
}

func (ac *AuthControllerImpl) AcceptInvite(c *fiber.Ctx) error {
    var body struct {
        Token    string `json:"token"`
        Name     string `json:"name"`
        Password string `json:"password"`
    }

    if err := c.BodyParser(&body); err != nil {
        return helpers.Error(c, 400, "invalid request")
    }

    user, err := ac.authService.AcceptInvite(c.Context(), body.Token, body.Name, body.Password)
    if err != nil {
        return helpers.Error(c, 400, err.Error())
    }

    user.PasswordHash = nil
    return helpers.Success(c, user)
}
