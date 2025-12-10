package service

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type ProjectServiceImpl struct {
	activityRepo repo.ActivityRepository
	projectRepo  repo.ProjectRepository
}

func NewProjectService(projectRepo repo.ProjectRepository, activityRepo repo.ActivityRepository) service.ProjectService {
	return &ProjectServiceImpl{
		activityRepo: activityRepo,
		projectRepo:  projectRepo,
	}
}

//
// ─────────────────────────────────────────────────────────────
//   HELPERS
// ─────────────────────────────────────────────────────────────
//

func (s *ProjectServiceImpl) ensureProjectAndTenant(ctx context.Context, customerID, projectID string) error {
	pr, err := s.projectRepo.GetByID(ctx, projectID, customerID)
	if err != nil || pr == nil {
		return errors.New("project not found or tenant mismatch")
	}
	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   CRUD
// ─────────────────────────────────────────────────────────────
//

func (s *ProjectServiceImpl) CreateProject(ctx context.Context, customerID, name, slug string) (*models.Project, error) {

	if strings.TrimSpace(name) == "" {
		return nil, errors.New("project name cannot be empty")
	}

	if strings.TrimSpace(slug) == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = strings.ReplaceAll(slug, " ", "-")

	existing, err := s.projectRepo.GetBySlug(ctx, slug, customerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("slug already exists for this customer")
	}

	p := &models.Project{
		ID:         uuid.NewString(),
		CustomerID: customerID,
		Name:       name,
		Slug:       slug,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProjectServiceImpl) GetProjects(ctx context.Context, customerID string) ([]models.Project, error) {
	return s.projectRepo.GetAll(ctx, customerID)
}

func (s *ProjectServiceImpl) GetProjectByID(ctx context.Context, id, customerID string) (*models.Project, error) {
	proj, err := s.projectRepo.GetByID(ctx, id, customerID)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, errors.New("project not found")
	}
	return proj, nil
}

func (s *ProjectServiceImpl) UpdateProject(ctx context.Context, id, customerID, name, slug string) (*models.Project, error) {
	proj, err := s.projectRepo.GetByID(ctx, id, customerID)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, errors.New("project not found")
	}

	if strings.TrimSpace(name) != "" {
		proj.Name = name
	}

	if strings.TrimSpace(slug) != "" {
		slug = strings.ToLower(strings.TrimSpace(slug))

		if slug != proj.Slug {
			existing, err := s.projectRepo.GetBySlug(ctx, slug, customerID)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.New("slug already exists for this customer")
			}
		}
		proj.Slug = slug
	}

	if err := s.projectRepo.Update(ctx, proj); err != nil {
		return nil, err
	}

	return proj, nil
}

func (s *ProjectServiceImpl) DeleteProject(ctx context.Context, id, customerID string) error {
	proj, err := s.projectRepo.GetByID(ctx, id, customerID)
	if err != nil {
		return err
	}
	if proj == nil {
		return errors.New("project not found")
	}

	return s.projectRepo.Delete(ctx, id, customerID)
}

//
// ─────────────────────────────────────────────────────────────
//   PROJECT ACTIVITY
// ─────────────────────────────────────────────────────────────
//

func (s *ProjectServiceImpl) ListProjectActivity(ctx context.Context, customerID, projectID string) ([]models.IssueActivity, error) {

	if err := s.ensureProjectAndTenant(ctx, customerID, projectID); err != nil {
		return nil, err
	}

	return s.activityRepo.ListByProject(ctx, projectID)
}
