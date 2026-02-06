package handler

import (
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	uploadService service.UploadService
}

func NewUploadHandler(svc service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: svc}
}

func (h *UploadHandler) GetSignedURL(c *fiber.Ctx) error {
	filename := c.Query("filename")
	contentType := c.Query("contentType")
	fileHash := c.Query("hash")

	if filename == "" || contentType == "" || fileHash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": false,
			"error":  "filename and contentType are required",
		})
	}

	result, err := h.uploadService.GetUploadInstructions(fileHash, filename, contentType)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to generate signed URL: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"result":  result,
	})
}
