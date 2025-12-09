package controllers

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	service "bugforge-backend/internal/service/interfaces"

	"github.com/gofiber/fiber/v2"
)

type ProjectControllerImpl struct {
	projectService service.ProjectService
}

func NewProjectController(s service.ProjectService) controller.ProjectController {
	return &ProjectControllerImpl{
		projectService: s,
	}
}


type CreateProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}


func (pc *ProjectControllerImpl) Create(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var body CreateProjectRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request")
	}

	project, err := pc.projectService.CreateProject(
		c.Context(),
		customerID.(string),
		body.Name,
		body.Slug,
	)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return helpers.Success(c, project)
}

// List all projects
// @Summary Get all projects
// @Description Returns all projects for the authenticated customer
// @Tags Projects
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /projects [get]
func (pc *ProjectControllerImpl) GetAll(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	projects, err := pc.projectService.GetProjects(c.Context(), customerID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return helpers.Success(c, projects)
}


func (pc *ProjectControllerImpl) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")

	proj, err := pc.projectService.GetProjectByID(c.Context(), id, customerID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusNotFound, err.Error())
	}

	return helpers.Success(c, proj)
}


func (pc *ProjectControllerImpl) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")

	var body UpdateProjectRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request")
	}

	proj, err := pc.projectService.UpdateProject(
		c.Context(),
		id,
		customerID.(string),
		body.Name,
		body.Slug,
	)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return helpers.Success(c, proj)
}


func (pc *ProjectControllerImpl) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	customerID := c.Locals("customer_id")

	if err := pc.projectService.DeleteProject(c.Context(), id, customerID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return helpers.Success(c, fiber.Map{"deleted": true})
}


func (pc *ProjectControllerImpl) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	customerID := c.Locals("customer_id")

	proj, err := pc.projectService.GetProjectByID(c.Context(), slug, customerID.(string))
	if err != nil {
		return helpers.Error(c, fiber.StatusNotFound, err.Error())
	}

	return helpers.Success(c, proj)
}


