package service

import (
	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	userRepo    repo.UserRepository
	projectRepo repo.ProjectRepository
	memberRepo  repo.ProjectMemberRepository
}

func NewUserService(userRepo repo.UserRepository, projectRepo repo.ProjectRepository, memberRepo repo.ProjectMemberRepository) service.UserService {
	return &UserServiceImpl{
		userRepo:    userRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, customerID, name, email, password, role string, assignedProjectIDs []string, defaultProjectID *string) (*models.User, error) {
	// basic validations
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		return nil, errors.New("name and email are required")
	}
	email = strings.ToLower(strings.TrimSpace(email))

	// email uniqueness within system
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	// verify assigned projects belong to same customer
	for _, pid := range assignedProjectIDs {
		pr, err := s.projectRepo.GetByID(ctx, pid, customerID)
		if err != nil {
			return nil, err
		}
		if pr == nil {
			return nil, errors.New("one or more assigned projects do not exist or do not belong to this customer")
		}
	}

	// default project must be one of assigned projects if provided
	if defaultProjectID != nil {
		found := false
		for _, pid := range assignedProjectIDs {
			if pid == *defaultProjectID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("default project must be one of assigned projects")
		}
	}

	var passwordHash *string
	if password != "" {
		bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = helpers.StrPtr(string(bs))
	} else {
		return nil, errors.New("password is required")
	}

	u := &models.User{
		ID:               uuid.NewString(),
		CustomerID:       customerID,
		Name:             helpers.StrPtr(name),
		Email:            email,
		PasswordHash:     passwordHash,
		Role:             role,
		DefaultProjectID: defaultProjectID,
	}


	// create user and assignments inside repo
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	if len(assignedProjectIDs) > 0 {
    	if err := s.memberRepo.SyncMembersForUser(ctx, u.ID, customerID, assignedProjectIDs); err != nil {
			// attempt cleanup
			_ = s.userRepo.Delete(ctx, u.ID, customerID)
			return nil, err
		}
		u.AssignedProjects = assignedProjectIDs
	}

	return u, nil
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id, customerID string) (*models.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil || u.CustomerID != customerID {
		return nil, errors.New("user not found")
	}

	pids, err := s.memberRepo.GetAssignedProjectIDsForUser(ctx, id)
    if err != nil {
        return nil, err
    }
    u.AssignedProjects = pids

	return u, nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email, customerID string) (*models.User, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil || u.CustomerID != customerID {
		return nil, nil
	}
	return u, nil
}

func (s *UserServiceImpl) GetAllByCustomer(ctx context.Context, customerID string) ([]models.User, error) {
	return s.userRepo.GetAllByCustomer(ctx, customerID)
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, id, customerID, name, email, password, role string, assignedProjectIDs []string, defaultProjectID *string) (*models.User, error) {
	u, err := s.GetByID(ctx, id, customerID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}

	// apply updates
	if strings.TrimSpace(name) != "" {
    u.Name = helpers.StrPtr(name)               // ← FIXED
	}

	if strings.TrimSpace(email) != "" {
		u.Email = strings.ToLower(strings.TrimSpace(email))
	}

	if strings.TrimSpace(role) != "" {
		u.Role = role
	}

	if password != "" {
		bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = helpers.StrPtr(string(bs))  // ← FIXED
	}
	// validate assigned projects belong to same customer
	if assignedProjectIDs != nil {
		// validate projects exist and belong to customer
		for _, pid := range assignedProjectIDs {
			pr, err := s.projectRepo.GetByID(ctx, pid, customerID)
			if err != nil {
				return nil, err
			}
			if pr == nil {
				return nil, errors.New("one or more assigned projects do not exist or do not belong to this customer")
			}
		}

		// Sync project_members (replace memberships)
		if err := s.memberRepo.SyncMembersForUser(ctx, u.ID, customerID, assignedProjectIDs); err != nil {
			return nil, err
		}

		// Update in-memory representation for response
		u.AssignedProjects = assignedProjectIDs
	}

	if defaultProjectID != nil {
		found := false
		for _, pid := range u.AssignedProjects {
			if pid == *defaultProjectID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("default project must be one of assigned projects")
		}
		u.DefaultProjectID = defaultProjectID
	}

	// persist update
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, id, customerID string) error {
	// ensure user exists in tenant
	u, err := s.GetByID(ctx, id, customerID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.DeleteInvitesByUserID(ctx, id); err != nil {
        return err
    }

	return s.userRepo.Delete(ctx, id, customerID)
}
