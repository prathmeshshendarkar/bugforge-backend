package routes

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"

	"github.com/gofiber/fiber/v2"
)

func ProjectRoutes(router fiber.Router, 
	pc controller.ProjectController, 
	ic controller.IssueController, 
	pmc controller.ProjectMemberController,
	labelCtrl controller.ProjectLabelController,
) {
	r := router.Group("/projects")

	// Project CRUD
	r.Post("/", pc.Create)
	r.Get("/", pc.GetAll)
	r.Get("/:id", pc.GetByID)
	r.Put("/:id", pc.Update)
	r.Delete("/:id", pc.Delete)

	// Project-specific issue listing
	r.Get("/:project_id/issues", ic.ListByProject)

	// ---- NEW: Project Members ----
	m := r.Group("/:project_id/members")
	
	m.Post("/invite", pmc.Invite)

	m.Get("/", pmc.List)
	m.Post("/", pmc.Add)
	m.Delete("/:user_id", pmc.Remove)

	// LABELS (project-level)
    r.Get("/:project_id/labels", labelCtrl.List)
    r.Post("/:project_id/labels", labelCtrl.Create)
    r.Patch("/:project_id/labels/:label_id", labelCtrl.Update)
    r.Delete("/:project_id/labels/:label_id", labelCtrl.Delete)
}
