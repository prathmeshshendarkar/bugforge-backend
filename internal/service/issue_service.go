package service

import (
	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	websocket "bugforge-backend/internal/websocket/comments"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IssueServiceImpl struct {
	issueRepo    repo.IssueRepository
	projectRepo  repo.ProjectRepository
	userRepo     repo.UserRepository
	commentRepo  repo.CommentRepository
	activityRepo repo.ActivityRepository
	activity     service.ActivityService
	commentHub   *websocket.CommentHub
	notifications service.NotificationService
}

func NewIssueService(
	issueRepo repo.IssueRepository,
	projectRepo repo.ProjectRepository,
	userRepo repo.UserRepository,
	commentRepo repo.CommentRepository,
	activityRepo repo.ActivityRepository,
	activitySvc service.ActivityService,
	commentHub *websocket.CommentHub,
	notifSvc service.NotificationService,
) service.IssueService {
	return &IssueServiceImpl{
		issueRepo:    issueRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
		commentRepo:  commentRepo,
		activityRepo: activityRepo,
		activity:     activitySvc,
		commentHub:   commentHub,
		notifications: notifSvc,
	}
}

//
// ─────────────────────────────────────────────────────────────
//   VALIDATION
// ─────────────────────────────────────────────────────────────
//

var (
	ErrInvalidStatus   = errors.New("invalid status")
	ErrInvalidPriority = errors.New("invalid priority")
)

var validStatuses = map[string]bool{
	"open": true, "in_progress": true, "resolved": true, "closed": true,
}

var validPriorities = map[string]bool{
	"low": true, "medium": true, "high": true, "critical": true,
}

//
// ─────────────────────────────────────────────────────────────
//   HELPERS
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) ensureIssueAndTenant(ctx context.Context, customerID, issueID string) (*models.Issue, error) {
	iss, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil || iss == nil {
		return nil, errors.New("issue not found")
	}

	pr, err := s.projectRepo.GetByID(ctx, iss.ProjectID, customerID)
	if err != nil || pr == nil {
		return nil, errors.New("tenant mismatch")
	}

	return iss, nil
}

func (s *IssueServiceImpl) ensureProjectAndTenant(ctx context.Context, customerID, projectID string) error {
	pr, err := s.projectRepo.GetByID(ctx, projectID, customerID)
	if err != nil || pr == nil {
		return errors.New("project not found")
	}
	return nil
}

func (s *IssueServiceImpl) notify(userID, title, message string, metadata map[string]interface{}) {
    b, _ := json.Marshal(metadata)

    _ = s.notifications.SendInApp(
        userID,
        title,
        message,
        string(b),
    )
}


//
// ─────────────────────────────────────────────────────────────
//   CORE ISSUE CRUD
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) CreateIssue(
	ctx context.Context,
	customerID, projectID, title, description, priority string,
	assignedTo *string,
	actorUserID string,
) (*models.Issue, error) {

	if err := s.ensureProjectAndTenant(ctx, customerID, projectID); err != nil {
		return nil, err
	}

	if title == "" {
		return nil, errors.New("title is required")
	}

	if !validPriorities[priority] {
		priority = "medium"
	}

	// Validate assignee
	if assignedTo != nil {
		u, err := s.userRepo.GetByID(ctx, *assignedTo)
		if err != nil || u == nil || u.CustomerID != customerID {
			return nil, errors.New("invalid assignee")
		}
	}

	issue := &models.Issue{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      "open",
		Priority:    priority,
		CreatedBy:   actorUserID,
		AssignedTo:  assignedTo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.issueRepo.Create(ctx, issue); err != nil {
		return nil, err
	}

	// Activity: issue created
	_ = s.activity.Log(ctx, issue.ID, &actorUserID, models.ActivityCreated, map[string]interface{}{
		"title": issue.Title,
	})

	// If assigned on create, log assignment as well
	if issue.AssignedTo != nil {
		_ = s.activity.Log(ctx, issue.ID, &actorUserID, models.ActivityAssigned, map[string]interface{}{
			"old": nil,
			"new": issue.AssignedTo,
		})
	}

	return issue, nil
}

func (s *IssueServiceImpl) ListAllIssues(ctx context.Context, customerID string) ([]models.IssueWithUser, error) {
	return s.issueRepo.ListAll(ctx, customerID)
}

func (s *IssueServiceImpl) GetIssue(ctx context.Context, customerID, issueID string) (*models.Issue, error) {
	return s.ensureIssueAndTenant(ctx, customerID, issueID)
}

func (s *IssueServiceImpl) ListIssuesByProject(ctx context.Context, projectID, customerID string, q url.Values) ([]models.IssueWithUser, error) {

	if err := s.ensureProjectAndTenant(ctx, customerID, projectID); err != nil {
		return nil, err
	}

	f := repo.IssueFilter{
		SortBy:    "created_at",
		Direction: "DESC",
		Limit:     20,
		Offset:    0,
	}

	if v := q.Get("status"); v != "" {
		f.Status = &v
	}
	if v := q.Get("priority"); v != "" {
		f.Priority = &v
	}
	if v := q.Get("assigned_to"); v != "" {
		f.AssignedTo = &v
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}

	if v := q.Get("limit"); v != "" {
		lim, _ := strconv.Atoi(v)
		if lim >= 5 && lim <= 200 {
			f.Limit = lim
		}
	}

	if v := q.Get("page"); v != "" {
		page, _ := strconv.Atoi(v)
		if page > 0 {
			f.Offset = (page - 1) * f.Limit
		}
	}

	if v := q.Get("sort"); v != "" {
		f.SortBy = v
	}

	if v := q.Get("direction"); v != "" {
		f.Direction = strings.ToUpper(v)
	}

	return s.issueRepo.ListByProject(ctx, projectID, f)
}

//
// ─────────────────────────────────────────────────────────────
//   UPDATE ISSUE
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) UpdateIssue(
	ctx context.Context,
	customerID, issueID string,
	title, description, status, priority string,
	assignedTo *string,
	actorUserID string,
) (*models.Issue, error) {

	i, err := s.ensureIssueAndTenant(ctx, customerID, issueID)
	if err != nil {
		return nil, err
	}

	// Keep old values for diffing
	oldStatus := i.Status
	oldAssigned := i.AssignedTo
	oldTitle := i.Title
	oldDescription := i.Description
	oldPriority := i.Priority

	// Apply updates only if provided
	if title != "" {
		i.Title = title
	}
	if description != "" {
		i.Description = description
	}
	if status != "" {
		if !validStatuses[status] {
			return nil, ErrInvalidStatus
		}
		i.Status = status
	}
	if priority != "" {
		if !validPriorities[priority] {
			return nil, ErrInvalidPriority
		}
		i.Priority = priority
	}

	if assignedTo != nil {
		u, err := s.userRepo.GetByID(ctx, *assignedTo)
		if err != nil || u == nil || u.CustomerID != customerID {
			return nil, errors.New("invalid assignee")
		}
		i.AssignedTo = assignedTo
	}

	i.UpdatedAt = time.Now()

	if err := s.issueRepo.Update(ctx, i); err != nil {
		return nil, err
	}

	// Activity logs (extended)
	// Status changed
	if status != "" && status != oldStatus {
		_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityStatusChanged, map[string]interface{}{
			"old": oldStatus,
			"new": status,
		})

		
		// Notify assigned user if exists
		if i.AssignedTo != nil {
			s.notify(*i.AssignedTo,
				"Issue Status Updated",
				fmt.Sprintf("Status changed to %s for issue %s", i.Status, i.Title),
				map[string]interface{}{
					"issue_id": issueID,
					"field":    "status",
					"old":      oldStatus,
					"new":      i.Status,
				},
			)
		}
	}

	// Assignment changed
	if (oldAssigned == nil && i.AssignedTo != nil) ||
		(oldAssigned != nil && i.AssignedTo == nil) ||
		(oldAssigned != nil && i.AssignedTo != nil && *oldAssigned != *i.AssignedTo) {

		_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityAssigned, map[string]interface{}{
			"old": oldAssigned,
			"new": i.AssignedTo,
		})

		fmt.Println(i.AssignedTo);

		if i.AssignedTo != nil {
			s.notify(*i.AssignedTo,
				"New Assignment",
				fmt.Sprintf("You were assigned to issue %s", i.Title),
				map[string]interface{}{
					"issue_id": issueID,
					"field":    "assigned_to",
					"old":      oldAssigned,
					"new":      i.AssignedTo,
				},
			)
		}

	}

	// Title changed
	if title != "" && title != oldTitle {
		_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityTitleUpdated, map[string]interface{}{
			"old": oldTitle,
			"new": title,
		})

		fmt.Println(i.AssignedTo);
		if i.AssignedTo != nil {
			s.notify(*i.AssignedTo,
				"Issue Title Updated",
				fmt.Sprintf("Issue renamed to %s", i.Title),
				map[string]interface{}{
					"issue_id": issueID,
					"field":    "title",
					"old":      oldTitle,
					"new":      i.Title,
				},
			)
		}
	}

	// Description changed
	if description != "" && description != oldDescription {
		_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityDescriptionUpdated, map[string]interface{}{
			"old": oldDescription,
			"new": description,
		})
	}

	// Priority changed
	if priority != "" && priority != oldPriority {
		_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityPriorityChanged, map[string]interface{}{
			"old": oldPriority,
			"new": priority,
		})

		fmt.Println("Priority Changed ", i.AssignedTo);

		if i.AssignedTo != nil {
			fmt.Println("Priority Changed ", i.AssignedTo);
			s.notify(*i.AssignedTo,
				"Priority Updated",
				fmt.Sprintf("Priority updated to %s", i.Priority),
				map[string]interface{}{
					"issue_id": issueID,
					"field":    "priority",
					"old":      oldPriority,
					"new":      i.Priority,
				},
			)
		}
	}

	return i, nil
}

//
// ─────────────────────────────────────────────────────────────
//   DELETE ISSUE
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) DeleteIssue(ctx context.Context, customerID, issueID, actorUserID string) error {

	i, err := s.ensureIssueAndTenant(ctx, customerID, issueID)
	if err != nil {
		return err
	}

	if err := s.issueRepo.Delete(ctx, issueID); err != nil {
		return err
	}

	_ = s.activity.Log(ctx, issueID, &actorUserID, models.ActivityDeleted, map[string]interface{}{
		"title": i.Title,
	})

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   DUE DATE
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) UpdateDueDate(ctx context.Context, customerID, issueID string, dueDate *time.Time, userID string) error {
	_, err := s.ensureIssueAndTenant(ctx, customerID, issueID)
	if err != nil {
		return err
	}

	// Save to DB
	if err := s.issueRepo.UpdateDueDate(ctx, issueID, dueDate); err != nil {
		return err
	}

	// Save activity
	_ = s.activity.Log(ctx, issueID, &userID, models.ActivityDueDateUpdated, map[string]interface{}{
		"due_date": dueDate,
	})

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   RELATIONS
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) AddRelation(ctx context.Context, customerID, issueID, relatedIssueID, relationType, userID string) error {

	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return err
	}

	rel := &models.IssueRelation{
		ID:             uuid.NewString(),
		IssueID:        issueID,
		RelatedIssueID: relatedIssueID,
		RelationType:   relationType,
		CreatedAt:      time.Now(),
	}

	if err := s.issueRepo.CreateRelation(ctx, rel); err != nil {
		return err
	}

	_ = s.activity.Log(ctx, issueID, &userID, models.ActivityRelationAdded, map[string]interface{}{
		"related_issue": relatedIssueID,
		"type":          relationType,
	})

	return nil
}

func (s *IssueServiceImpl) ListRelations(ctx context.Context, customerID, issueID string) ([]models.IssueRelation, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListRelations(ctx, issueID)
}

func (s *IssueServiceImpl) DeleteRelation(ctx context.Context, customerID, relationID, userID string) error {
	// We cannot tenant-check directly from relation — repo does not expose GetRelationByID
	// So simply delete and trust controller-level validation.
	// Optionally log a relation deletion if you can fetch relation details.
	_ = s.activity.Log(ctx, relationID, &userID, models.ActivityRelationAdded, map[string]interface{}{
		"relation_id": relationID,
	})
	return s.issueRepo.DeleteRelation(ctx, relationID)
}

//
// ─────────────────────────────────────────────────────────────
//   COMMENTS (rich text + edit/delete + updated_at)
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) CreateComment(
	ctx context.Context,
	customerID, issueID, userID, body string,
) (*models.IssueComment, error) {

	if body == "" {
		return nil, errors.New("comment cannot be empty")
	}

	// Validate tenant & issue
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}

	// Sanitize rich text HTML
	bodyHTML := helpers.SanitizeHTML(body)

	c := &models.IssueComment{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		UserID:    userID,
		Body:      body,
		BodyHTML:  helpers.StrPtr(bodyHTML),
		CreatedAt: time.Now(),
		UpdatedAt: nil,
	}

	// DB: insert comment
	if err := s.issueRepo.CreateComment(ctx, c); err != nil {
		return nil, err
	}

	// Save activity: "commented"
	_ = s.activity.Log(ctx, issueID, &userID, models.ActivityCommented, map[string]interface{}{
		"comment_id": c.ID,
	})

	// Notify issue assignee for new comment
	iss, _ := s.issueRepo.GetByID(ctx, issueID)
	if iss != nil && iss.AssignedTo != nil && *iss.AssignedTo != userID {
		s.notify(*iss.AssignedTo,
			"New Comment",
			fmt.Sprintf("A new comment was added to: %s", iss.Title),
			map[string]interface{}{
				"issue_id": issueID,
				"comment_id": c.ID,
			},
		)
	}

	usernames := helpers.ExtractMentions(body)

	for _, uname := range usernames {
		u, err := s.userRepo.GetByUsername(ctx, uname)
		if err != nil || u == nil {
			continue
		}
		if u.CustomerID != customerID {
			continue
		}

		mentionEvt := websocket.CommentEvent{
			Type:    "mention",
			IssueID: issueID,
			ActorID: userID,
			Payload: map[string]interface{}{
				"mentioned_user_id": u.ID,
				"comment_id":        c.ID,
			},
		}

		s.commentHub.BroadcastToUser(u.ID, mentionEvt)

		_ = s.activity.Log(ctx, issueID, &userID, models.ActivityMentioned, map[string]interface{}{
			"mentioned_user": u.ID,
		})

		s.notify(u.ID,
			"You were mentioned",
			fmt.Sprintf("You were mentioned in issue %s", issueID),
			map[string]interface{}{
				"issue_id": issueID,
				"comment_id": c.ID,
				"mentioned_by": userID,
			},
		)

	}

	// Fetch user info for WS + API consistency
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil {
		c.AuthorName = user.Name
		c.AuthorEmail = &user.Email
	}

	evt := websocket.CommentEvent{
		Type:    "comment_created",
		IssueID: issueID,
		ActorID: userID,
		Payload: c,
	}

	b, _ := json.Marshal(evt)
	s.commentHub.GetRoom(issueID).Broadcast(b)

	return c, nil
}

func (s *IssueServiceImpl) UpdateComment(
	ctx context.Context,
	customerID, commentID, userID, body string,
) (*models.IssueComment, error) {

	if body == "" {
		return nil, errors.New("comment cannot be empty")
	}

	// Fetch existing comment
	c, err := s.issueRepo.GetCommentByID(ctx, commentID)
	if err != nil || c == nil {
		return nil, errors.New("comment not found")
	}

	// Tenant validation
	if _, err := s.ensureIssueAndTenant(ctx, customerID, c.IssueID); err != nil {
		return nil, err
	}

	// Ownership check
	if c.UserID != userID {
		return nil, errors.New("cannot edit others' comments")
	}

	// Keep old body for diffing
	oldBody := c.Body

	// Update fields
	c.Body = body
	c.BodyHTML = helpers.StrPtr(helpers.SanitizeHTML(body))

	now := time.Now()
	c.UpdatedAt = &now

	// Save DB
	if err := s.issueRepo.UpdateComment(ctx, c); err != nil {
		return nil, err
	}

	// Log comment_updated activity
	_ = s.activity.Log(ctx, c.IssueID, &userID, models.ActivityCommented, map[string]interface{}{
		"comment_id": c.ID,
		"old":        oldBody,
		"new":        body,
	})

	// PROCESS MENTIONS
	usernames := helpers.ExtractMentions(body)

	for _, uname := range usernames {
		u, err := s.userRepo.GetByUsername(ctx, uname)
		if err != nil || u == nil {
			continue
		}
		if u.CustomerID != customerID {
			continue
		}

		// notify via WS
		mentionEvt := websocket.CommentEvent{
			Type:    "mention",
			IssueID: c.IssueID,
			ActorID: userID,
			Payload: map[string]interface{}{
				"mentioned_user_id": u.ID,
				"comment_id":        c.ID,
			},
		}

		s.commentHub.BroadcastToUser(u.ID, mentionEvt)

		// log activity for mention
		_ = s.activity.Log(ctx, c.IssueID, &userID, models.ActivityMentioned, map[string]interface{}{
			"comment_id":     c.ID,
			"mentioned_user": u.ID,
		})
	}

	// WS BROADCAST: comment_updated
	evt := websocket.CommentEvent{
		Type:    "comment_updated",
		IssueID: c.IssueID,
		ActorID: userID,
		Payload: c,
	}

	b, _ := json.Marshal(evt)
	s.commentHub.GetRoom(c.IssueID).Broadcast(b)

	return c, nil
}

func (s *IssueServiceImpl) DeleteComment(
	ctx context.Context,
	customerID, commentID, userID string,
) error {

	c, err := s.issueRepo.GetCommentByID(ctx, commentID)
	if err != nil || c == nil {
		return errors.New("comment not found")
	}

	// Ownership check
	if c.UserID != userID {
		return errors.New("cannot delete others' comments")
	}

	// Tenant validation
	if _, err := s.ensureIssueAndTenant(ctx, customerID, c.IssueID); err != nil {
		return err
	}

	evt := websocket.CommentEvent{
		Type:    "comment_deleted",
		IssueID: c.IssueID,
		ActorID: userID,
		Payload: map[string]string{"id": commentID},
	}
	b, _ := json.Marshal(evt)
	s.commentHub.GetRoom(c.IssueID).Broadcast(b)

	// Log deletion
	_ = s.activity.Log(ctx, c.IssueID, &userID, models.ActivityCommentDeleted, map[string]interface{}{
		"comment_id": commentID,
	})

	return s.issueRepo.DeleteComment(ctx, commentID)
}

func (s *IssueServiceImpl) ListComments(
	ctx context.Context,
	customerID, issueID string,
) ([]models.IssueComment, error) {

	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}

	return s.issueRepo.ListCommentsByIssue(ctx, issueID)
}

//
// ─────────────────────────────────────────────────────────────
//   ATTACHMENTS
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) AddAttachment(ctx context.Context, customerID, issueID, userID string, att *models.IssueAttachment) error {

	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return err
	}

	att.ID = uuid.NewString()
	att.IssueID = issueID
	att.UserID = userID
	att.CreatedAt = time.Now()

	if err := s.issueRepo.CreateAttachment(ctx, att); err != nil {
		return err
	}

	_ = s.activity.Log(ctx, issueID, &userID, models.ActivityAttachmentAdded, map[string]interface{}{
		"filename": att.Filename,
	})

	return nil
}

func (s *IssueServiceImpl) ListAttachments(ctx context.Context, customerID, issueID string) ([]models.IssueAttachment, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListAttachmentsByIssue(ctx, issueID)
}

func (s *IssueServiceImpl) DeleteAttachment(ctx context.Context, customerID, attachmentID, userID string) error {
	// Repo cannot validate tenant from attachment alone — assuming controller checks issueID.
	// Optionally you can load the attachment to include filename in activity metadata.
	return s.issueRepo.DeleteAttachment(ctx, attachmentID)
}

//
// ─────────────────────────────────────────────────────────────
//   CHECKLISTS + ITEMS
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) CreateChecklist(ctx context.Context, customerID, issueID, title, userID string) (*models.Checklist, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}

	cl := &models.Checklist{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		Title:     title,
		CreatedAt: time.Now(),
	}

	if err := s.issueRepo.CreateChecklist(ctx, cl); err != nil {
		return nil, err
	}

	_ = s.activity.Log(ctx, issueID, &userID, models.ActivityChecklistCreated, map[string]interface{}{
		"checklist_id": cl.ID,
		"title":        cl.Title,
	})

	return cl, nil
}

func (s *IssueServiceImpl) ListChecklists(ctx context.Context, customerID, issueID string) ([]models.Checklist, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListChecklistsByIssue(ctx, issueID)
}

func (s *IssueServiceImpl) CreateChecklistItem(ctx context.Context, customerID, checklistID, content string, userID string) (*models.ChecklistItem, error) {

	// We cannot tenant-check checklist directly without joining; assume controller handles mapping.

	item := &models.ChecklistItem{
		ID:          uuid.NewString(),
		ChecklistID: checklistID,
		Content:     content,
		Done:        false,
		OrderIndex:  0,
	}

	if err := s.issueRepo.CreateChecklistItem(ctx, item); err != nil {
		return nil, err
	}

	_ = s.activity.Log(ctx, item.ChecklistID, &userID, models.ActivityChecklistItemAdded, map[string]interface{}{
		"item_id": item.ID,
		"content": item.Content,
	})

	return item, nil
}

func (s *IssueServiceImpl) UpdateChecklistItem(ctx context.Context, customerID, itemID string, content string, done bool, userID string) (*models.ChecklistItem, error) {

	// Repo missing GetChecklistItem — skipping tenant check. If available, fetch it first for old values.
	item := &models.ChecklistItem{
		ID:         itemID,
		Content:    content,
		Done:       done,
		OrderIndex: 0, // not updated
	}

	err := s.issueRepo.UpdateChecklistItem(ctx, item)
	if err != nil {
		return nil, err
	}

	_ = s.activity.Log(ctx, item.ChecklistID, &userID, models.ActivityChecklistItemUpdated, map[string]interface{}{
		"item_id": item.ID,
		"content": item.Content,
		"done":    item.Done,
	})

	return item, nil
}

func (s *IssueServiceImpl) DeleteChecklist(ctx context.Context, customerID, checklistID string) error {
	// TODO: ideally validate checklist belongs to issue & tenant

	if err := s.issueRepo.DeleteChecklist(ctx, checklistID); err != nil {
		return err
	}

	_ = s.activity.Log(ctx, checklistID, nil, models.ActivityChecklistDeleted, map[string]interface{}{
		"checklist_id": checklistID,
	})

	return nil
}

func (s *IssueServiceImpl) DeleteChecklistItem(ctx context.Context, customerID, itemID string) error {
	// Optionally load checklistID for metadata
	_ = s.activity.Log(ctx, itemID, nil, models.ActivityChecklistItemDeleted, map[string]interface{}{
		"item_id": itemID,
	})
	return s.issueRepo.DeleteChecklistItem(ctx, itemID)
}

func (s *IssueServiceImpl) ReorderChecklistItems(
	ctx context.Context,
	customerID, checklistID string,
	order []models.ChecklistItem,
) error {
	if err := s.issueRepo.ReorderChecklistItems(ctx, checklistID, order); err != nil {
		return err
	}

	// Log reorder with simple metadata (order size)
	_ = s.activity.Log(ctx, checklistID, nil, models.ActivityChecklistReordered, map[string]interface{}{
		"count": len(order),
	})

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   SUBTASKS
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) CreateSubtask(
	ctx context.Context,
	customerID, issueID, title string,
	description *string,
	assignedTo *string,
	dueDate *time.Time,
	userID string,
) (*models.Subtask, error) {

	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}

	sub := &models.Subtask{
		ID:            uuid.NewString(),
		ParentIssueID: issueID,
		Title:         title,
		Description:   description,
		Status:        "open",
		AssignedTo:    assignedTo,
		DueDate:       dueDate,
		OrderIndex:    0,
		CreatedAt:     time.Now(),
	}

	if err := s.issueRepo.CreateSubtask(ctx, sub); err != nil {
		return nil, err
	}

	_ = s.activity.Log(ctx, issueID, &userID, models.ActivitySubtaskCreated, map[string]interface{}{
		"subtask_id": sub.ID,
		"title":      sub.Title,
	})

	return sub, nil
}

func (s *IssueServiceImpl) UpdateSubtask(
	ctx context.Context,
	customerID, subtaskID string,
	title string,
	description *string,
	status string,
	assignedTo *string,
	dueDate *time.Time,
	userID string,
) (*models.Subtask, error) {

	//  Get existing subtask first
	existing, err := s.issueRepo.GetSubtaskByID(ctx, subtaskID)
	if err != nil || existing == nil {
		return nil, errors.New("subtask not found")
	}

	// Keep old values for diffing
	oldTitle := existing.Title
	oldDescription := existing.Description
	oldStatus := existing.Status
	oldAssigned := existing.AssignedTo
	oldDue := existing.DueDate

	//  Update fields only if provided
	if title != "" {
		existing.Title = title
	}
	if description != nil {
		existing.Description = description
	}
	if status != "" {
		existing.Status = status
	}
	if assignedTo != nil {
		existing.AssignedTo = assignedTo
	}
	if dueDate != nil {
		existing.DueDate = dueDate
	}

	//  Save updated row
	if err := s.issueRepo.UpdateSubtask(ctx, existing); err != nil {
		return nil, err
	}

	//  Fetch correct final row from DB
	updated, _ := s.issueRepo.GetSubtaskByID(ctx, subtaskID)

	// Log changes (only fields that changed)
	if title != "" && title != oldTitle {
		_ = s.activity.Log(ctx, updated.ParentIssueID, &userID, models.ActivitySubtaskUpdated, map[string]interface{}{
			"subtask_id": subtaskID,
			"field":      "title",
			"old":        oldTitle,
			"new":        title,
		})
	}
	if description != nil && (oldDescription == nil || *oldDescription != *description) {
		_ = s.activity.Log(ctx, updated.ParentIssueID, &userID, models.ActivitySubtaskUpdated, map[string]interface{}{
			"subtask_id": subtaskID,
			"field":      "description",
		})
	}
	if status != "" && status != oldStatus {
		_ = s.activity.Log(ctx, updated.ParentIssueID, &userID, models.ActivitySubtaskUpdated, map[string]interface{}{
			"subtask_id": subtaskID,
			"field":      "status",
			"old":        oldStatus,
			"new":        status,
		})

	}
	if (oldAssigned == nil && existing.AssignedTo != nil) ||
		(oldAssigned != nil && existing.AssignedTo == nil) ||
		(oldAssigned != nil && existing.AssignedTo != nil && *oldAssigned != *existing.AssignedTo) {
		_ = s.activity.Log(ctx, updated.ParentIssueID, &userID, models.ActivitySubtaskUpdated, map[string]interface{}{
			"subtask_id": subtaskID,
			"field":      "assigned",
			"old":        oldAssigned,
			"new":        existing.AssignedTo,
		})
	}
	if (oldDue == nil && existing.DueDate != nil) ||
		(oldDue != nil && existing.DueDate == nil) ||
		(oldDue != nil && existing.DueDate != nil && !oldDue.Equal(*existing.DueDate)) {
		_ = s.activity.Log(ctx, updated.ParentIssueID, &userID, models.ActivitySubtaskUpdated, map[string]interface{}{
			"subtask_id": subtaskID,
			"field":      "due_date",
		})
	}

	return updated, nil
}

func (s *IssueServiceImpl) ListSubtasks(ctx context.Context, customerID, issueID string) ([]models.Subtask, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListSubtasksByParent(ctx, issueID)
}

func (s *IssueServiceImpl) DeleteSubtask(ctx context.Context, customerID, subtaskID string) error {
	// Optionally fetch subtask for parent issue id
	sub, _ := s.issueRepo.GetSubtaskByID(ctx, subtaskID)
	if err := s.issueRepo.DeleteSubtask(ctx, subtaskID); err != nil {
		return err
	}
	if sub != nil {
		_ = s.activity.Log(ctx, sub.ParentIssueID, nil, models.ActivitySubtaskDeleted, map[string]interface{}{
			"subtask_id": subtaskID,
		})
	}
	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   ACTIVITY
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) ListActivity(ctx context.Context, customerID, issueID string) ([]models.IssueActivity, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.activityRepo.ListByIssue(ctx, issueID)
}
