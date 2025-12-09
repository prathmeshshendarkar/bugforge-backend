package service

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"errors"
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
}

func NewIssueService(
	issueRepo repo.IssueRepository,
	projectRepo repo.ProjectRepository,
	userRepo repo.UserRepository,
	commentRepo repo.CommentRepository,
	activityRepo repo.ActivityRepository,
) service.IssueService {
	return &IssueServiceImpl{
		issueRepo:    issueRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
		commentRepo:  commentRepo,
		activityRepo: activityRepo,
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

	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   issue.ID,
		UserID:    &actorUserID,
		Action:    "created",
		Metadata:  map[string]interface{}{"title": title},
		CreatedAt: time.Now(),
	})

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

	oldStatus := i.Status
	oldAssigned := i.AssignedTo

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

	// Activity logs
	if status != "" && status != oldStatus {
		_ = s.activityRepo.Create(ctx, &models.IssueActivity{
			ID:        uuid.NewString(),
			IssueID:   issueID,
			UserID:    &actorUserID,
			Action:    "status_changed",
			Metadata:  map[string]interface{}{"old": oldStatus, "new": status},
			CreatedAt: time.Now(),
		})
	}

	if (oldAssigned == nil && i.AssignedTo != nil) ||
		(oldAssigned != nil && i.AssignedTo == nil) ||
		(oldAssigned != nil && i.AssignedTo != nil && *oldAssigned != *i.AssignedTo) {

		_ = s.activityRepo.Create(ctx, &models.IssueActivity{
			ID:        uuid.NewString(),
			IssueID:   issueID,
			UserID:    &actorUserID,
			Action:    "assigned",
			Metadata:  map[string]interface{}{"old": oldAssigned, "new": i.AssignedTo},
			CreatedAt: time.Now(),
		})
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

	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   i.ID,
		UserID:    &actorUserID,
		Action:    "deleted",
		CreatedAt: time.Now(),
	})

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//   DUE DATE
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) UpdateDueDate(ctx context.Context, customerID, issueID string, dueDate *time.Time, userID string) error {
	i, err := s.ensureIssueAndTenant(ctx, customerID, issueID)
	if err != nil {
		return err
	}

	// Save to DB
	if err := s.issueRepo.UpdateDueDate(ctx, issueID, dueDate); err != nil {
		return err
	}

	// Save activity
	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   i.ID,
		UserID:    &userID,
		Action:    "due_date_updated",
		Metadata:  map[string]interface{}{"due_date": dueDate},
		CreatedAt: time.Now(),
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

	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		UserID:    &userID,
		Action:    "relation_added",
		Metadata:  map[string]interface{}{"related_issue": relatedIssueID, "type": relationType},
		CreatedAt: time.Now(),
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
	return s.issueRepo.DeleteRelation(ctx, relationID)
}

//
// ─────────────────────────────────────────────────────────────
//   COMMENTS (add edit/delete)
// ─────────────────────────────────────────────────────────────
//

func (s *IssueServiceImpl) CreateComment(ctx context.Context, customerID, issueID, userID, body string) (*models.IssueComment, error) {
	if body == "" {
		return nil, errors.New("comment cannot be empty")
	}

	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}

	c := &models.IssueComment{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		UserID:    userID,
		Body:      body,
		CreatedAt: time.Now(),
	}

	if err := s.issueRepo.CreateComment(ctx, c); err != nil {
		return nil, err
	}

	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		UserID:    &userID,
		Action:    "commented",
		Metadata:  map[string]interface{}{"comment_id": c.ID},
		CreatedAt: time.Now(),
	})

	return c, nil
}

func (s *IssueServiceImpl) UpdateComment(ctx context.Context, customerID, commentID, userID, body string) (*models.IssueComment, error) {

	if body == "" {
		return nil, errors.New("comment cannot be empty")
	}

	c, err := s.issueRepo.GetCommentByID(ctx, commentID)
	if err != nil || c == nil {
		return nil, errors.New("comment not found")
	}

	_, err = s.ensureIssueAndTenant(ctx, customerID, c.IssueID)
	if err != nil {
		return nil, err
	}

	if c.UserID != userID {
		return nil, errors.New("cannot edit others' comments")
	}

	c.Body = body

	if err := s.issueRepo.UpdateComment(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *IssueServiceImpl) DeleteComment(ctx context.Context, customerID, commentID, userID string) error {

	c, err := s.issueRepo.GetCommentByID(ctx, commentID)
	if err != nil || c == nil {
		return errors.New("comment not found")
	}

	if c.UserID != userID {
		return errors.New("cannot delete others' comments")
	}

	_, err = s.ensureIssueAndTenant(ctx, customerID, c.IssueID)
	if err != nil {
		return err
	}

	return s.issueRepo.DeleteComment(ctx, commentID)
}

func (s *IssueServiceImpl) ListComments(ctx context.Context, customerID, issueID string) ([]models.IssueComment, error) {
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

	_ = s.activityRepo.Create(ctx, &models.IssueActivity{
		ID:        uuid.NewString(),
		IssueID:   issueID,
		UserID:    &userID,
		Action:    "attachment_added",
		Metadata:  map[string]interface{}{"filename": att.Filename},
		CreatedAt: time.Now(),
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
		ID:         uuid.NewString(),
		ChecklistID: checklistID,
		Content:    content,
		Done:       false,
		OrderIndex: 0,
	}

	if err := s.issueRepo.CreateChecklistItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *IssueServiceImpl) UpdateChecklistItem(ctx context.Context, customerID, itemID string, content string, done bool, userID string) (*models.ChecklistItem, error) {

	// Repo missing GetChecklistItem — skipping tenant check.
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

	return item, nil
}

func (s *IssueServiceImpl) DeleteChecklist(ctx context.Context, customerID, checklistID string) error {
	// TODO: ideally validate checklist belongs to issue & tenant

	return s.issueRepo.DeleteChecklist(ctx, checklistID)
}


func (s *IssueServiceImpl) DeleteChecklistItem(ctx context.Context, customerID, itemID string) error {
	return s.issueRepo.DeleteChecklistItem(ctx, itemID)
}


func (s *IssueServiceImpl) ReorderChecklistItems(
	ctx context.Context,
	customerID, checklistID string,
	order []models.ChecklistItem,
) error {
	return s.issueRepo.ReorderChecklistItems(ctx, checklistID, order)
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

    // 🔥 Get existing subtask first
    existing, err := s.issueRepo.GetSubtaskByID(ctx, subtaskID)
    if err != nil || existing == nil {
        return nil, errors.New("subtask not found")
    }

    // 🔥 Update fields only if provided
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

    // 🔥 Save updated row
    if err := s.issueRepo.UpdateSubtask(ctx, existing); err != nil {
        return nil, err
    }

    // 🔥 Fetch correct final row from DB
    updated, _ := s.issueRepo.GetSubtaskByID(ctx, subtaskID)
    return updated, nil
}


func (s *IssueServiceImpl) ListSubtasks(ctx context.Context, customerID, issueID string) ([]models.Subtask, error) {
	if _, err := s.ensureIssueAndTenant(ctx, customerID, issueID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListSubtasksByParent(ctx, issueID)
}

func (s *IssueServiceImpl) DeleteSubtask(ctx context.Context, customerID, subtaskID string) error {
    return s.issueRepo.DeleteSubtask(ctx, subtaskID)
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
