package controllers

import (
	"bugforge-backend/internal/http/controllers/interfaces"
	"bugforge-backend/internal/http/helpers"
	"bugforge-backend/internal/models"
	service "bugforge-backend/internal/service/interfaces"
	"context"

	"github.com/gofiber/fiber/v2"
)

type IssueAttachmentControllerImpl struct {
	svc service.IssueService
}

func NewIssueAttachmentController(s service.IssueService) interfaces.IssueAttachmentController {
	return &IssueAttachmentControllerImpl{svc: s}
}

// Upload: minimal multipart handling. For production, prefer pre-signed S3 uploads or streaming to object storage.
// This handler will extract basic metadata and pass it to the service which will persist metadata.
// TODO: integrate S3 presign or direct upload and set the attachment URL/key accordingly.
func (ia *IssueAttachmentControllerImpl) Upload(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")

	// accept multipart file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		// allow client to send attachment metadata as JSON instead
		var att models.IssueAttachment
		if err := c.BodyParser(&att); err != nil {
			return helpers.Error(c, fiber.StatusBadRequest, "no file or metadata provided")
		}
		if err := ia.svc.AddAttachment(context.Background(), customerID.(string), issueID, userID.(string), &att); err != nil {
			return helpers.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return helpers.Success(c, att)
	}

	// Basic metadata from uploaded file (we don't store bytes here)
	size := fileHeader.Size
	filename := fileHeader.Filename
	contentType := fileHeader.Header.Get("Content-Type")

	att := &models.IssueAttachment{
		URL:        "", // TODO: set to generated URL after upload to S3
		Key:        filename,
		Filename:   filename,
		ContentType: &contentType,
		Size:       int64(size),
	}

	if err := ia.svc.AddAttachment(context.Background(), customerID.(string), issueID, userID.(string), att); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return helpers.Success(c, att)
}

func (ia *IssueAttachmentControllerImpl) Delete(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	userID := c.Locals("user_id")
	if customerID == nil || userID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	attachmentID := c.Params("attachment_id")
	if err := ia.svc.DeleteAttachment(context.Background(), customerID.(string), attachmentID, userID.(string)); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, fiber.Map{"deleted": true})
}

func (ia *IssueAttachmentControllerImpl) List(c *fiber.Ctx) error {
	customerID := c.Locals("customer_id")
	if customerID == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}
	issueID := c.Params("id")

	out, err := ia.svc.ListAttachments(context.Background(), customerID.(string), issueID)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return helpers.Success(c, out)
}
