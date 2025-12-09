package interfaces

import "github.com/gofiber/fiber/v2"

// Comments
type IssueCommentController interface {
	Create(c *fiber.Ctx) error
	List(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

// Relations
type IssueRelationController interface {
	Add(c *fiber.Ctx) error
	List(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

// Attachments
type IssueAttachmentController interface {
	Upload(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	List(c *fiber.Ctx) error
}

// Checklists
type IssueChecklistController interface {
	Create(c *fiber.Ctx) error
	AddItem(c *fiber.Ctx) error
	UpdateItem(c *fiber.Ctx) error
	List(c *fiber.Ctx) error
	DeleteChecklist(c *fiber.Ctx) error
	DeleteItem(c *fiber.Ctx) error
	ReorderItems(c *fiber.Ctx) error
}

// Subtasks
type IssueSubtaskController interface {
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	List(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}
