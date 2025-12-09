package auth

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
    UserID      string   `json:"user_id"`
    CustomerID  string   `json:"customer_id"`
    Roles       []string `json:"roles"`
    ClientIDs   []string `json:"client_ids,omitempty"`
    ProjectIDs  []string `json:"project_ids,omitempty"`
    AccessLevel string   `json:"access_level"`

    jwt.RegisteredClaims
}
