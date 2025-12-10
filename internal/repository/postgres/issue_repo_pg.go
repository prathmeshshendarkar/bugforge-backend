package postgres

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IssueRepoPG struct {
	db *pgxpool.Pool
}

func NewIssueRepository(db *pgxpool.Pool) repo.IssueRepository {
	return &IssueRepoPG{db: db}
}

//
// ─────────────────────────────────────────────────────────────
//   CORE ISSUE CRUD
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) Create(ctx context.Context, i *models.Issue) error {
	query := `
		INSERT INTO issues (id, project_id, title, description, status, priority, created_by, assigned_to, due_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())

	`
	if i.ID == "" {
		i.ID = uuid.NewString()
	}
	_, err := r.db.Exec(ctx, query,
		i.ID, i.ProjectID, i.Title, i.Description,
		i.Status, i.Priority, i.CreatedBy, i.AssignedTo,
		i.DueDate,
	)

	return err
}

func (r *IssueRepoPG) ListAll(ctx context.Context, customerID string) ([]models.IssueWithUser, error) {
	query := `
        SELECT i.id, i.project_id, i.title, i.description, i.status, i.priority,
               i.created_by, cu.email AS created_by_email, cu.name AS created_by_name,
               i.assigned_to, au.email AS assigned_to_email, au.name AS assigned_to_name,
               i.created_at, i.updated_at
        FROM issues i
        LEFT JOIN users cu ON cu.id = i.created_by
        LEFT JOIN users au ON au.id = i.assigned_to
        WHERE cu.customer_id = $1
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueWithUser

	for rows.Next() {
		var i models.IssueWithUser
		err := rows.Scan(
			&i.ID, &i.ProjectID, &i.Title, &i.Description, &i.Status, &i.Priority,
			&i.CreatedBy, &i.CreatedByEmail, &i.CreatedByName,
			&i.AssignedTo, &i.AssignedToEmail, &i.AssignedToName,
			&i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}

	return out, nil
}

func (r *IssueRepoPG) GetByID(ctx context.Context, id string) (*models.Issue, error) {
	query := `
		SELECT id, project_id, title, description, status, priority, created_by,
			assigned_to, due_date, created_at, updated_at
		FROM issues WHERE id=$1 LIMIT 1

	`

	var i models.Issue

	err := r.db.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.ProjectID, &i.Title, &i.Description, &i.Status,
		&i.Priority, &i.CreatedBy, &i.AssignedTo, &i.DueDate,
		&i.CreatedAt, &i.UpdatedAt,

	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *IssueRepoPG) ListByProject(ctx context.Context, projectID string, f repo.IssueFilter) ([]models.IssueWithUser, error) {
	baseQuery := `
        SELECT 
            i.id, i.project_id, i.title, i.description, i.status, i.priority,
            i.created_by, cb.email AS created_by_email, cb.name AS created_by_name,
            i.assigned_to, ab.email AS assigned_to_email, ab.name AS assigned_to_name,
            i.created_at, i.updated_at
        FROM issues i
        LEFT JOIN users cb ON cb.id = i.created_by
        LEFT JOIN users ab ON ab.id = i.assigned_to
        WHERE i.project_id = $1
	`

	params := []interface{}{projectID}
	idx := 2

	if f.Status != nil {
		baseQuery += fmt.Sprintf(" AND i.status = $%d", idx)
		params = append(params, *f.Status)
		idx++
	}
	if f.Priority != nil {
		baseQuery += fmt.Sprintf(" AND i.priority = $%d", idx)
		params = append(params, *f.Priority)
		idx++
	}
	if f.AssignedTo != nil {
		baseQuery += fmt.Sprintf(" AND i.assigned_to = $%d", idx)
		params = append(params, *f.AssignedTo)
		idx++
	}
	if f.Search != nil {
		baseQuery += fmt.Sprintf(" AND (i.title ILIKE $%d OR i.description ILIKE $%d)", idx, idx)
		params = append(params, "%"+*f.Search+"%")
		idx++
	}

	baseQuery += fmt.Sprintf(" ORDER BY i.%s %s", f.SortBy, f.Direction)

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	params = append(params, f.Limit, f.Offset)

	rows, err := r.db.Query(ctx, baseQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueWithUser
	for rows.Next() {
		var i models.IssueWithUser
		err := rows.Scan(
			&i.ID, &i.ProjectID, &i.Title, &i.Description,
			&i.Status, &i.Priority,
			&i.CreatedBy, &i.CreatedByEmail, &i.CreatedByName,
			&i.AssignedTo, &i.AssignedToEmail, &i.AssignedToName,
			&i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}

	return out, nil
}

func (r *IssueRepoPG) Update(ctx context.Context, i *models.Issue) error {
	query := `
		UPDATE issues
		SET title=$1, description=$2, status=$3, priority=$4, assigned_to=$5, due_date=$6, updated_at=NOW()
		WHERE id=$7

	`
	_, err := r.db.Exec(ctx, query,
		i.Title, i.Description, i.Status, i.Priority, i.AssignedTo, i.DueDate, i.ID,
	)
	return err
}

func (r *IssueRepoPG) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM issues WHERE id=$1`, id)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//   COMMENTS
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) CreateComment(ctx context.Context, c *models.IssueComment) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}

	query := `
		INSERT INTO issue_comments (id, issue_id, user_id, body, created_at)
		VALUES ($1,$2,$3,$4,NOW())
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.IssueID, c.UserID, c.Body)
	return err
}

func (r *IssueRepoPG) GetCommentByID(ctx context.Context, id string) (*models.IssueComment, error) {
	query := `
		SELECT id, issue_id, user_id, body, created_at
		FROM issue_comments WHERE id = $1
	`

	var c models.IssueComment

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.IssueID, &c.UserID, &c.Body, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *IssueRepoPG) UpdateComment(ctx context.Context, c *models.IssueComment) error {
	query := `
		UPDATE issue_comments
		SET body=$1
		WHERE id=$2
	`
	_, err := r.db.Exec(ctx, query, c.Body, c.ID)
	return err
}

func (r *IssueRepoPG) DeleteComment(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM issue_comments WHERE id=$1`, id)
	return err
}

func (r *IssueRepoPG) ListCommentsByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error) {
	query := `
		SELECT 
			c.id,
			c.issue_id,
			c.user_id,
			c.body,
			c.body_html,
			c.created_at,
			c.updated_at,
			u.name AS author_name,
			u.email AS author_email
		FROM issue_comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.issue_id = $1
		ORDER BY c.created_at ASC;
	`

	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueComment

	for rows.Next() {
		var c models.IssueComment

		err := rows.Scan(
			&c.ID,
			&c.IssueID,
			&c.UserID,
			&c.Body,
			&c.BodyHTML,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.AuthorName,
			&c.AuthorEmail,
		)

		if err != nil {
			return nil, err
		}

		out = append(out, c)
	}

	return out, nil
}

//
// ─────────────────────────────────────────────────────────────
//   ATTACHMENTS
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) CreateAttachment(ctx context.Context, a *models.IssueAttachment) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}

	query := `
		INSERT INTO issue_attachments (id, issue_id, user_id, url, key, filename, content_type, size, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`

	_, err := r.db.Exec(ctx, query,
		a.ID, a.IssueID, a.UserID, a.URL, a.Key,
		a.Filename, a.ContentType, a.Size,
	)

	return err
}

func (r *IssueRepoPG) ListAttachmentsByIssue(ctx context.Context, issueID string) ([]models.IssueAttachment, error) {
	query := `
		SELECT id, issue_id, user_id, url, key, filename, content_type, size, created_at
		FROM issue_attachments
		WHERE issue_id=$1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueAttachment

	for rows.Next() {
		var a models.IssueAttachment
		err := rows.Scan(
			&a.ID, &a.IssueID, &a.UserID, &a.URL, &a.Key,
			&a.Filename, &a.ContentType, &a.Size, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	return out, nil
}

func (r *IssueRepoPG) DeleteAttachment(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM issue_attachments WHERE id=$1`, id)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//   CHECKLISTS
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) CreateChecklist(ctx context.Context, cl *models.Checklist) error {
	if cl.ID == "" {
		cl.ID = uuid.NewString()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO checklists (id, issue_id, title, created_at)
		VALUES ($1,$2,$3,NOW())
	`, cl.ID, cl.IssueID, cl.Title)

	return err
}

func (r *IssueRepoPG) CreateChecklistItem(ctx context.Context, it *models.ChecklistItem) error {
	if it.ID == "" {
		it.ID = uuid.NewString()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO checklist_items (id, checklist_id, content, done, order_index)
		VALUES ($1,$2,$3,$4,$5)
	`,
		it.ID, it.ChecklistID, it.Content, it.Done, it.OrderIndex,
	)

	return err
}

func (r *IssueRepoPG) UpdateChecklistItem(ctx context.Context, it *models.ChecklistItem) error {
	_, err := r.db.Exec(ctx, `
		UPDATE checklist_items
		SET content=$1, done=$2, order_index=$3
		WHERE id=$4
	`,
		it.Content, it.Done, it.OrderIndex, it.ID,
	)

	return err
}

func (r *IssueRepoPG) ListChecklistsByIssue(ctx context.Context, issueID string) ([]models.Checklist, error) {
	query := `
		SELECT 
		    cl.id AS checklist_id,
		    cl.issue_id,
		    cl.title,
		    cl.created_at,
		    it.id AS item_id,
		    it.content,
		    it.done,
		    it.order_index
		FROM checklists cl
		LEFT JOIN checklist_items it 
		    ON it.checklist_id = cl.id
		WHERE cl.issue_id = $1
		ORDER BY cl.created_at ASC, it.order_index ASC
	`

	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	outMap := map[string]*models.Checklist{}

	for rows.Next() {
		var (
			clID, issID, title string
			createdAt          time.Time
			itemID             *string
			content            *string
			done               *bool
			orderIndex         *int
		)

		err := rows.Scan(&clID, &issID, &title, &createdAt, &itemID, &content, &done, &orderIndex)
		if err != nil {
			return nil, err
		}

		if outMap[clID] == nil {
			outMap[clID] = &models.Checklist{
				ID:        clID,
				IssueID:   issID,
				Title:     title,
				CreatedAt: createdAt,
				Items:     []models.ChecklistItem{},
			}
		}

		// Item exists
		if itemID != nil {
			outMap[clID].Items = append(outMap[clID].Items, models.ChecklistItem{
				ID:         *itemID,
				ChecklistID: clID,
				Content:    *content,
				Done:       *done,
				OrderIndex: *orderIndex,
			})
		}
	}

	var output []models.Checklist
	for _, v := range outMap {
		output = append(output, *v)
	}

	return output, nil
}

func (r *IssueRepoPG) DeleteChecklist(ctx context.Context, checklistID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM checklists WHERE id=$1`, checklistID)
	return err
}

func (r *IssueRepoPG) DeleteChecklistItem(ctx context.Context, itemID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM checklist_items WHERE id=$1`, itemID)
	return err
}

func (r *IssueRepoPG) ReorderChecklistItems(ctx context.Context, checklistID string, items []models.ChecklistItem) error {
	batch := &pgx.Batch{}

	for _, it := range items {
		batch.Queue(`
			UPDATE checklist_items 
			SET order_index=$1 
			WHERE id=$2 AND checklist_id=$3
		`, it.OrderIndex, it.ID, checklistID)
	}

	_, err := r.db.SendBatch(ctx, batch).Exec()
	return err
}


//
// ─────────────────────────────────────────────────────────────
//   SUBTASKS
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) CreateSubtask(ctx context.Context, s *models.Subtask) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO subtasks (id, parent_issue_id, title, description, status, assigned_to, due_date, order_index, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`,
		s.ID, s.ParentIssueID, s.Title, s.Description, s.Status, s.AssignedTo,
		s.DueDate, s.OrderIndex,
	)

	return err
}

func (r *IssueRepoPG) GetSubtaskByID(ctx context.Context, id string) (*models.Subtask, error) {
    query := `
        SELECT id, parent_issue_id, title, description, status,
               assigned_to, due_date, order_index, created_at
        FROM subtasks
        WHERE id = $1
    `

    var s models.Subtask
    err := r.db.QueryRow(ctx, query, id).Scan(
        &s.ID, &s.ParentIssueID, &s.Title, &s.Description, &s.Status,
        &s.AssignedTo, &s.DueDate, &s.OrderIndex, &s.CreatedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, nil
    }
    return &s, err
}

func (r *IssueRepoPG) UpdateSubtask(ctx context.Context, s *models.Subtask) error {
	_, err := r.db.Exec(ctx, `
		UPDATE subtasks
		SET title=$1, description=$2, status=$3, assigned_to=$4, due_date=$5, order_index=$6
		WHERE id=$7
	`,
		s.Title, s.Description, s.Status, s.AssignedTo,
		s.DueDate, s.OrderIndex, s.ID,
	)

	return err
}

func (r *IssueRepoPG) ListSubtasksByParent(ctx context.Context, parentIssueID string) ([]models.Subtask, error) {
	query := `
		SELECT id, parent_issue_id, title, description, status, assigned_to, due_date, order_index, created_at
		FROM subtasks
		WHERE parent_issue_id=$1
		ORDER BY order_index ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, parentIssueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Subtask

	for rows.Next() {
		var s models.Subtask
		err := rows.Scan(
			&s.ID, &s.ParentIssueID, &s.Title, &s.Description, &s.Status,
			&s.AssignedTo, &s.DueDate, &s.OrderIndex, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, nil
}

func (r *IssueRepoPG) DeleteSubtask(ctx context.Context, id string) error {
    _, err := r.db.Exec(ctx, `DELETE FROM subtasks WHERE id=$1`, id)
    return err
}


//
// ─────────────────────────────────────────────────────────────
//   ISSUE RELATIONS
// ─────────────────────────────────────────────────────────────
//

func (r *IssueRepoPG) CreateRelation(ctx context.Context, rel *models.IssueRelation) error {
	if rel.ID == "" {
		rel.ID = uuid.NewString()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO issue_relations (id, issue_id, related_issue_id, relation_type, created_at)
		VALUES ($1,$2,$3,$4,NOW())
	`,
		rel.ID, rel.IssueID, rel.RelatedIssueID, rel.RelationType,
	)

	return err
}

func (r *IssueRepoPG) ListRelations(ctx context.Context, issueID string) ([]models.IssueRelation, error) {
	query := `
		SELECT id, issue_id, related_issue_id, relation_type, created_at
		FROM issue_relations
		WHERE issue_id=$1
	`

	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IssueRelation

	for rows.Next() {
		var rel models.IssueRelation
		err := rows.Scan(
			&rel.ID, &rel.IssueID, &rel.RelatedIssueID, &rel.RelationType, &rel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, rel)
	}

	return out, nil
}

func (r *IssueRepoPG) DeleteRelation(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM issue_relations WHERE id=$1`, id)
	return err
}

func (r *IssueRepoPG) UpdateDueDate(ctx context.Context, issueID string, dueDate *time.Time) error {
	query := `
		UPDATE issues 
		SET due_date=$1, updated_at=NOW()
		WHERE id=$2
	`
	_, err := r.db.Exec(ctx, query, dueDate, issueID)
	return err
}
