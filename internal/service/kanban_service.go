package service

import (
	"context"
	"errors"

	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
)

// KanbanServiceImpl contains the repositories it needs.
// Note: projectMemberRepo is used to check if a user belongs to a project.
type KanbanServiceImpl struct {
	issueRepo         repo.IssueRepository
	projectRepo       repo.ProjectRepository       // kept in case you need project CRUD in future
	projectMemberRepo repo.ProjectMemberRepository
	kanbanRepo        repo.KanbanRepository
}

func NewKanbanService(
	issueRepo repo.IssueRepository,
	projectRepo repo.ProjectRepository,
	projectMemberRepo repo.ProjectMemberRepository,
	kanbanRepo repo.KanbanRepository,
) *KanbanServiceImpl {
	return &KanbanServiceImpl{
		issueRepo:         issueRepo,
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		kanbanRepo:        kanbanRepo,
	}
}


func (s *KanbanServiceImpl) GetBoard(projectID string, userID string) (*models.KanbanBoard, error) {
    ctx := context.Background()

    // validate membership
    isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("forbidden")
    }

    cols, err := s.kanbanRepo.GetColumnsWithCards(ctx, projectID)
    if err != nil {
        return nil, err
    }

    board := &models.KanbanBoard{
        Columns: cols,
    }

    return board, nil
}

//
// ---------------------------------------------------------------
// CREATE COLUMN
// ---------------------------------------------------------------
//

func (s *KanbanServiceImpl) CreateColumn(
	projectID string,
	name string,
	userID string,
) (*models.KanbanColumn, error) {

	ctx := context.Background()

	// Validate project membership via projectMemberRepo
	isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("forbidden")
	}

	// Determine next order
	order, err := s.kanbanRepo.GetNextColumnOrder(ctx, projectID)
	if err != nil {
		return nil, err
	}

	col := &models.KanbanColumn{
		ID:        models.NewUUID(),
		ProjectID: projectID,
		Name:      name,
		Order:     order,
	}

	if err := s.kanbanRepo.CreateColumn(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}

//
// ---------------------------------------------------------------
// CREATE CARD
// ---------------------------------------------------------------
//

func (s *KanbanServiceImpl) CreateCard(
	projectID string,
	columnID string,
	title string,
	description string,
	userID string,
) (*models.Issue, error) {

	ctx := context.Background()

	// Validate project membership via projectMemberRepo
	isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("forbidden")
	}

	// Determine next order inside the column
	nextOrder, err := s.kanbanRepo.GetNextOrder(ctx, columnID)
	if err != nil {
		return nil, err
	}

	card := &models.Issue{
		ID:          models.NewUUID(),
		ProjectID:   projectID,
		ColumnID:    columnID,
		Order:       nextOrder,
		Title:       title,
		Description: description,
		Status:      "todo",
		Priority:    "medium",
		CreatedBy:   userID,
	}

	if err := s.kanbanRepo.CreateCard(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

//
// ---------------------------------------------------------------
// MOVE CARD
// ---------------------------------------------------------------
//

func (s *KanbanServiceImpl) MoveCard(
    cardID string,
    toColumnID string,
    newOrder int,
    userID string,
) (*models.Issue, string, error) {

    ctx := context.Background()

    // 1. Fetch existing card
    card, err := s.kanbanRepo.GetCardByID(ctx, cardID)
    if err != nil {
        return nil, "", errors.New("card_not_found")
    }

    fromColumnID := card.ColumnID
    oldOrder := card.Order

    // 2. Validate membership
    isMember, err := s.projectMemberRepo.IsMember(ctx, card.ProjectID, userID)
    if err != nil {
        return nil, "", err
    }
    if !isMember {
        return nil, "", errors.New("forbidden")
    }

    // 3. Perform move inside TX
    err = s.kanbanRepo.Tx(ctx, func(tx repo.KanbanRepository) error {

        // Moving to a different column
        if fromColumnID != toColumnID {

            if err := tx.ShiftOrdersDown(ctx, fromColumnID, oldOrder); err != nil {
                return err
            }

            if err := tx.ShiftOrders(ctx, toColumnID, newOrder); err != nil {
                return err
            }

        } else {
            // Reordering inside SAME column
            if newOrder < oldOrder {
                if err := tx.ShiftRangeUp(ctx, fromColumnID, newOrder, oldOrder); err != nil {
                    return err
                }
            } else if newOrder > oldOrder {
                if err := tx.ShiftRangeDown(ctx, fromColumnID, newOrder, oldOrder); err != nil {
                    return err
                }
            }
        }

        // Update card
        card.ColumnID = toColumnID
        card.Order = newOrder

        return tx.UpdateCardPosition(ctx, card)
    })

    if err != nil {
        return nil, "", err
    }

    return card, fromColumnID, nil
}

func (s *KanbanServiceImpl) ReorderColumn(
    projectID string,
    columnID string,
    newOrder int,
    userID string,
) error {

    ctx := context.Background()

    // Validate membership
    isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("forbidden")
    }

    // Load all columns
    cols, err := s.kanbanRepo.GetColumnsWithCards(ctx, projectID)
    if err != nil {
        return err
    }

    // Find target column + reorder logic
    var oldOrder int
    for _, col := range cols {
        if col.ID == columnID {
            oldOrder = col.Order
            break
        }
    }

    if oldOrder == 0 {
        return errors.New("column_not_found")
    }

    // Execute DB updates inside TX
    return s.kanbanRepo.Tx(ctx, func(tx repo.KanbanRepository) error {

        // Reorder logic:
        if newOrder < oldOrder {
            // Shift columns DOWN (order + 1)
            return tx.ReorderColumnsShiftUp(ctx, projectID, newOrder, oldOrder, columnID)
        }

        if newOrder > oldOrder {
            // Shift columns UP (order - 1)
            return tx.ReorderColumnsShiftDown(ctx, projectID, oldOrder, newOrder, columnID)
        }

        return nil
    })
}

func (s *KanbanServiceImpl) RenameColumn(projectID, columnID, newName, userID string) (*models.KanbanColumn, error) {
    ctx := context.Background()

    isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("forbidden")
    }

    col := &models.KanbanColumn{
        ID:        columnID,
        ProjectID: projectID,
        Name:      newName,
    }

    err = s.kanbanRepo.UpdateColumnName(ctx, columnID, newName)
    if err != nil {
        return nil, err
    }

    return col, nil
}

func (s *KanbanServiceImpl) DeleteColumn(projectID, columnID, userID string) error {
    ctx := context.Background()

    // Validate membership
    isMember, err := s.projectMemberRepo.IsMember(ctx, projectID, userID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("forbidden")
    }

    // Run inside transaction
    return s.kanbanRepo.Tx(ctx, func(tx repo.KanbanRepository) error {

        // 1. Delete cards in this column
        if err := tx.DeleteCardsByColumn(ctx, columnID); err != nil {
            return err
        }

        // 2. Delete column
        if err := tx.DeleteColumn(ctx, columnID); err != nil {
            return err
        }

        return nil
    })
}

func (s *KanbanServiceImpl) DeleteCard(cardID, userID string) (*models.Issue, error) {
    ctx := context.Background()

    // 1. Fetch card
    card, err := s.kanbanRepo.GetCardByID(ctx, cardID)
    if err != nil {
        return nil, errors.New("card_not_found")
    }

    // 2. Validate membership
    isMember, err := s.projectMemberRepo.IsMember(ctx, card.ProjectID, userID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("forbidden")
    }

    columnID := card.ColumnID
    oldOrder := card.Order

    // 3. Delete inside transaction
    err = s.kanbanRepo.Tx(ctx, func(tx repo.KanbanRepository) error {

        // Shift down other cards in the same column
        if err := tx.ShiftOrdersDown(ctx, columnID, oldOrder); err != nil {
            return err
        }

        // Delete the card
        return tx.DeleteCard(ctx, cardID)
    })

    if err != nil {
        return nil, err
    }

    return card, nil
}
