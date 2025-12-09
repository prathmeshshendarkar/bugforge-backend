package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	svc "bugforge-backend/internal/service/interfaces"

	"github.com/google/uuid"
)

type ProjectMemberServiceImpl struct {
	projectRepo repo.ProjectRepository
	userRepo    repo.UserRepository
	memberRepo  repo.ProjectMemberRepository
}

func NewProjectMemberService(
	projectRepo repo.ProjectRepository,
	userRepo repo.UserRepository,
	memberRepo repo.ProjectMemberRepository,
) svc.ProjectMemberService {
	return &ProjectMemberServiceImpl{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		memberRepo:  memberRepo,
	}
}

func (s *ProjectMemberServiceImpl) AddMember(ctx context.Context, projectID, customerID, userID string) error {
	// Validate project belongs to customer
	project, err := s.projectRepo.GetByID(ctx, projectID, customerID)
	if err != nil || project == nil {
		return errors.New("project not found")
	}

	// Validate user belongs to customer
	user, err := s.userRepo.GetByID(ctx, userID)
	fmt.Println(customerID, user);
	if err != nil || user == nil || user.CustomerID != customerID {
		return errors.New("user does not belong to this customer")
	}

	return s.memberRepo.AddMember(ctx, projectID, userID)
}

func (s *ProjectMemberServiceImpl) RemoveMember(ctx context.Context, projectID, customerID, userID string) error {
	// No need to validate project/user again
	return s.memberRepo.RemoveMember(ctx, projectID, userID)
}

func (s *ProjectMemberServiceImpl) ListMembers(ctx context.Context, projectID, customerID string) ([]models.User, error) {
	return s.memberRepo.ListMembers(ctx, projectID, customerID)
}

func (s *ProjectMemberServiceImpl) Invite(
    ctx context.Context,
    projectID, customerID, email, role string,
) error {

    // Validate project belongs to customer
    project, err := s.projectRepo.GetByID(ctx, projectID, customerID)
    if err != nil || project == nil {
        return errors.New("project not found")
    }

    // Try fetching the user by email
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return err
    }

    var userID string

    if user == nil {
        // ----------- User does not exist → Create pending user -----------
        tempUser := &models.User{
            ID:               uuid.NewString(),
            CustomerID:       customerID,
            Email:            email,
            Role:             role,
            IsPending:        true, // NEW FIELD
        }

        if err := s.userRepo.CreatePending(ctx, tempUser); err != nil {
            return err
        }

        // Generate invite token
        token := uuid.NewString()

        if err := s.userRepo.SaveInviteToken(ctx, tempUser.ID, token); err != nil {
            return err
        }

        // Send email async or log for now
        fmt.Println("SEND INVITE EMAIL TO:", email, "TOKEN:", token)

		inviteURL := fmt.Sprintf(
			"%s/accept-invite?token=%s",
			os.Getenv("FRONTEND_URL"),
			token,
		)

		fmt.Println(inviteURL);
		
		helpers.SendEmail(email,
			"You're invited to join a project",
			fmt.Sprintf("Click the link to activate your account: %s", inviteURL),
		)


        userID = tempUser.ID
    } else {
        // ----------- User exists -----------
        if user.CustomerID != customerID {
            return errors.New("email belongs to another customer")
        }

        userID = user.ID
    }

    // Add to project members
    return s.memberRepo.AddMember(ctx, projectID, userID)
}
