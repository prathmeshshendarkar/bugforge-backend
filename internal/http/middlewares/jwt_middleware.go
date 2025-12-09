package middleware

import (
	"bugforge-backend/internal/auth"
	"bugforge-backend/internal/http/helpers"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {

		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return helpers.Error(c, fiber.StatusUnauthorized, "Missing Authorization header")
		}

		// Remove "Bearer "
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.ParseWithClaims(tokenString, &auth.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid token")
		}

		claims, ok := token.Claims.(*auth.JWTClaims)
		if !ok {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid token claims")
		}

		// set locals
		c.Locals("user_id", claims.UserID)
		c.Locals("customer_id", claims.CustomerID)
		c.Locals("roles", claims.Roles)
		c.Locals("client_ids", claims.ClientIDs)
		c.Locals("project_ids", claims.ProjectIDs)
		c.Locals("access_level", claims.AccessLevel)

		return c.Next()
	}
}
