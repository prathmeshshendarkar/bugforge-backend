package service

import (
	"bugforge-backend/internal/auth"
	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"errors"
	"os"
	"strings"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	userRepo repo.UserRepository
}

func NewAuthService(userRepo repo.UserRepository) service.AuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	// password must exist
	if user.PasswordHash == nil {
		return nil, "", errors.New("invalid credentials")
	}

	// compare hash
	if bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// build claims (keep client_ids & project_ids as empty slices for now)
	claims := auth.JWTClaims{
		UserID:     user.ID,
		CustomerID: user.CustomerID,
		Roles:      []string{user.Role}, // wrap single role
		ClientIDs:  []string{},          // keep empty for now (populate later if needed)
		ProjectIDs: user.AssignedProjects,
		AccessLevel: "tenant",           // semantic label you can change later

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(48 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// Optionally set ID, Issuer, Subject if you want
		},
	}

	// sign token using env var JWT_SECRET
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// fallback or fail fast — better to fail so you don't issue unsigned tokens
		return nil, "", errors.New("jwt secret not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, "", err
	}

	// hide password before returning
	user.PasswordHash = nil

	return user, signedToken, nil
}

func (s *AuthServiceImpl) AcceptInvite(ctx context.Context, token, name, password string) (*models.User, error) {
    if strings.TrimSpace(name) == "" || strings.TrimSpace(password) == "" {
        return nil, errors.New("name and password required")
    }

    // find user by invite token
    user, err := s.userRepo.GetByInviteToken(ctx, token)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("invalid or expired invite token")
    }

    // hash password
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // update user
    user.Name = helpers.StrPtr(name)
    user.PasswordHash = helpers.StrPtr(string(hashed))
    user.IsPending = false

    if err := s.userRepo.MarkInviteAccepted(ctx, user.ID); err != nil {
        return nil, err
    }

    if err := s.userRepo.Update(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
