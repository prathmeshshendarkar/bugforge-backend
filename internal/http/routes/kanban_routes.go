package routes

import (
	"encoding/json"

	"bugforge-backend/internal/service"
	ws "bugforge-backend/internal/websocket"

	"github.com/gofiber/fiber/v2"
)

func RegisterKanbanRoutes(router fiber.Router, kanbanService *service.KanbanServiceImpl, hub *ws.Hub) {

    // GET BOARD
    router.Get("/projects/:projectID/kanban", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        userID := c.Locals("user_id").(string)

        board, err := kanbanService.GetBoard(projectID, userID)
        if err != nil {
            return err
        }

        return c.JSON(board)
    })

    // CREATE COLUMN
    router.Post("/projects/:projectID/columns", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        userID := c.Locals("user_id").(string)

        var body struct{ Name string }
        c.BodyParser(&body)

        col, err := kanbanService.CreateColumn(projectID, body.Name, userID)
        if err != nil {
            return err
        }

        // ---> WS BROADCAST
        room := hub.GetRoom(projectID)
        evt := map[string]any{
            "type":      "column_created",
            "projectID": projectID,
            "column":    col,
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(col)
    })

    // CREATE CARD
    router.Post("/projects/:projectID/columns/:columnID/cards", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        columnID := c.Params("columnID")
        userID := c.Locals("user_id").(string)

        var body struct {
            Title       string
            Description string
        }
        c.BodyParser(&body)

        card, err := kanbanService.CreateCard(projectID, columnID, body.Title, body.Description, userID)
        if err != nil {
            return err
        }

        // ---> WS BROADCAST
        room := hub.GetRoom(projectID)
        evt := map[string]any{
            "type":      "card_created",
            "projectID": projectID,
            "card":      card,
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(card)
    })

    // MOVE CARD
    router.Patch("/kanban/cards/:cardID/move", func(c *fiber.Ctx) error {
        cardID := c.Params("cardID")
        userID := c.Locals("user_id").(string)

        var body struct {
            ColumnID string `json:"columnId"`
            Order    int    `json:"order"`
        }
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        updatedCard, fromColumnID, err := kanbanService.MoveCard(cardID, body.ColumnID, body.Order, userID)
        if err != nil {
            return err
        }

        // Broadcast WS event
        room := hub.GetRoom(updatedCard.ProjectID)
        evt := map[string]any{
            "type": "card_moved",
            "projectID": updatedCard.ProjectID,
            "payload": map[string]any{
                "card_id":     updatedCard.ID,
                "from_column": fromColumnID,
                "to_column":   updatedCard.ColumnID,
                "new_order":   updatedCard.Order,
            },
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(updatedCard)
    })

    // REORDER COLUMNS
    router.Patch("/projects/:projectID/columns/reorder", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        userID := c.Locals("user_id").(string)

        var body struct {
            ColumnID string `json:"column_id"`
            NewOrder int    `json:"new_order"`
        }
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        // Call service
        err := kanbanService.ReorderColumn(projectID, body.ColumnID, body.NewOrder, userID)
        if err != nil {
            return err
        }

        // WS broadcast
        room := hub.GetRoom(projectID)
        evt := map[string]any{
            "type":      "column_reordered",
            "projectID": projectID,
            "payload": map[string]any{
                "column_id": body.ColumnID,
                "new_order": body.NewOrder,
            },
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(fiber.Map{"status": "ok"})
    })

    // RENAME COLUMN
    router.Patch("/projects/:projectID/columns/:columnID", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        columnID := c.Params("columnID")
        userID := c.Locals("user_id").(string)

        var body struct {
            Name string `json:"name"`
        }
        if err := c.BodyParser(&body); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid body")
        }

        updatedCol, err := kanbanService.RenameColumn(projectID, columnID, body.Name, userID)
        if err != nil {
            return err
        }

        // WS BROADCAST
        room := hub.GetRoom(projectID)
        evt := map[string]any{
            "type": "column_renamed",
            "payload": map[string]any{
                "column_id": columnID,
                "name":      body.Name,
            },
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(updatedCol)
    })


    // DELETE COLUMN (and all cards inside it)
    router.Delete("/projects/:projectID/columns/:columnID", func(c *fiber.Ctx) error {
        projectID := c.Params("projectID")
        columnID := c.Params("columnID")
        userID := c.Locals("user_id").(string)

        err := kanbanService.DeleteColumn(projectID, columnID, userID)
        if err != nil {
            return err
        }

        // WS BROADCAST
        room := hub.GetRoom(projectID)
        evt := map[string]any{
            "type": "column_deleted",
            "payload": map[string]any{
                "column_id": columnID,
            },
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(fiber.Map{"status": "ok"})
    })

    // DELETE CARD
    router.Delete("/kanban/cards/:cardID", func(c *fiber.Ctx) error {
        cardID := c.Params("cardID")
        userID := c.Locals("user_id").(string)

        deletedCard, err := kanbanService.DeleteCard(cardID, userID)
        if err != nil {
            return err
        }

        // WS BROADCAST
        room := hub.GetRoom(deletedCard.ProjectID)
        evt := map[string]any{
            "type":       "card_deleted",
            "project_id": deletedCard.ProjectID,
            "payload": map[string]any{
                "card_id":   deletedCard.ID,
                "column_id": deletedCard.ColumnID,
            },
        }
        b, _ := json.Marshal(evt)
        room.Broadcast(b)

        return c.JSON(fiber.Map{"success": true})
    })

}
