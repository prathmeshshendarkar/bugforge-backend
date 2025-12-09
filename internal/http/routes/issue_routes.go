package routes

import (
	controllers "bugforge-backend/internal/http/controllers/interfaces"

	"github.com/gofiber/fiber/v2"
)

func IssueRoutes(
	router fiber.Router,
	issueCtrl controllers.IssueController,
	commentCtrl controllers.IssueCommentController,
	relationCtrl controllers.IssueRelationController,
	attachmentCtrl controllers.IssueAttachmentController,
	checklistCtrl controllers.IssueChecklistController,
	subtaskCtrl controllers.IssueSubtaskController,
) {

	r := router.Group("/issues")

	// List by project MUST BE FIRST (avoid collision with /:id)
	r.Get("/project/:project_id", issueCtrl.ListByProject)

	
	// DueDate
	r.Patch("/:id/due-date", issueCtrl.UpdateDueDate)

	// Issues
	r.Post("/", issueCtrl.Create)
	r.Get("/", issueCtrl.ListAll)
	r.Get("/:id", issueCtrl.Get)
	r.Patch("/:id", issueCtrl.Update)
	r.Delete("/:id", issueCtrl.Delete)

	// Activity — SAFE here (after /:id but before other wildcards)
	r.Get("/:id/activity", issueCtrl.ListActivity)

	// Comments
	r.Post("/:id/comments", commentCtrl.Create)
	r.Get("/:id/comments", commentCtrl.List)
	r.Patch("/:id/comments/:comment_id", commentCtrl.Update)
	r.Delete("/:id/comments/:comment_id", commentCtrl.Delete)

	// Relations
	r.Post("/:id/relations", relationCtrl.Add)
	r.Get("/:id/relations", relationCtrl.List)
	r.Delete("/:id/relations/:relation_id", relationCtrl.Delete)

	// Attachments
	r.Post("/:id/attachments", attachmentCtrl.Upload)
	r.Delete("/:id/attachments/:attachment_id", attachmentCtrl.Delete)
	r.Get("/:id/attachments", attachmentCtrl.List)

	// Checklists
	r.Post("/:id/checklists", checklistCtrl.Create)
	r.Get("/:id/checklists", checklistCtrl.List)

	// Items (CRUD)
	r.Post("/:id/checklists/:checklist_id/items", checklistCtrl.AddItem)
	r.Patch("/:id/checklists/items/:item_id", checklistCtrl.UpdateItem)
	r.Delete("/:id/checklists/items/:item_id", checklistCtrl.DeleteItem)

	// Delete entire checklist
	r.Delete("/:id/checklists/:checklist_id", checklistCtrl.DeleteChecklist)

	// Reorder
	r.Post("/:id/checklists/:checklist_id/reorder", checklistCtrl.ReorderItems)

	// Subtasks
	r.Post("/:id/subtasks", subtaskCtrl.Create)
	r.Patch("/:id/subtasks/:subtask_id", subtaskCtrl.Update)
	r.Get("/:id/subtasks", subtaskCtrl.List)
	r.Delete("/:id/subtasks/:subtask_id", subtaskCtrl.Delete)

}
