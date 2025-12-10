package models

//
// ─────────────────────────────────────────────────────────────
//   ISSUE CORE EVENTS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivityCreated        = "created"
	ActivityDeleted        = "deleted"

	// Updated fields
	ActivityTitleUpdated       = "title_updated"
	ActivityDescriptionUpdated = "description_updated"
	ActivityPriorityChanged    = "priority_changed"
	ActivityStatusChanged      = "status_changed"
	ActivityAssigned           = "assigned"

	ActivityDueDateUpdated = "due_date_updated"
)

//
// ─────────────────────────────────────────────────────────────
//   COMMENTS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivityCommented       = "commented"        // create or update content
	ActivityCommentDeleted  = "comment_deleted"
	ActivityMentioned       = "mentioned"        // @mentions
)

//
// ─────────────────────────────────────────────────────────────
//   ATTACHMENTS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivityAttachmentAdded = "attachment_added"
)

//
// ─────────────────────────────────────────────────────────────
//   RELATIONS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivityRelationAdded = "relation_added"
	// Optional future: ActivityRelationDeleted = "relation_deleted"
)

//
// ─────────────────────────────────────────────────────────────
//   CHECKLISTS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivityChecklistCreated   = "checklist_created"
	ActivityChecklistDeleted   = "checklist_deleted"
	ActivityChecklistReordered = "checklist_reordered"

	ActivityChecklistItemAdded    = "checklist_item_added"
	ActivityChecklistItemUpdated  = "checklist_item_updated"
	ActivityChecklistItemDeleted  = "checklist_item_deleted"
)

//
// ─────────────────────────────────────────────────────────────
//   SUBTASKS
// ─────────────────────────────────────────────────────────────
//

const (
	ActivitySubtaskCreated = "subtask_created"
	ActivitySubtaskUpdated = "subtask_updated"
	ActivitySubtaskDeleted = "subtask_deleted"
)
