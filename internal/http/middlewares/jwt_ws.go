package middleware

import (
	"bugforge-backend/internal/auth"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtectedWebSocket() fiber.Handler {
    return func(c *fiber.Ctx) error {

        tokenString := c.Query("token")
		fmt.Println("WebSocket Token:", tokenString)
        if tokenString == "" {
            return fiber.ErrUnauthorized
        }

        token, err := jwt.ParseWithClaims(tokenString, &auth.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            return fiber.ErrUnauthorized
        }

        claims := token.Claims.(*auth.JWTClaims)

        // store locals
        c.Locals("user_id", claims.UserID)
        c.Locals("customer_id", claims.CustomerID)
        c.Locals("roles", claims.Roles)
        c.Locals("client_ids", claims.ClientIDs)
        c.Locals("project_ids", claims.ProjectIDs)
        c.Locals("access_level", claims.AccessLevel)

        return c.Next()
    }
}
