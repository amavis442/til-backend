package handler

import (
	"errors"
	"strconv"

	"github.com/amavis442/til-backend/internal/til"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
)

type TilHandler struct {
	service     til.Service
	userService user.Service
}

func NewTilHandler(s til.Service, u user.Service) *TilHandler {
	return &TilHandler{
		service:     s,
		userService: u,
	}
}

func (h *TilHandler) List(c *fiber.Ctx) error {
	tils, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch TILs"})
	}
	return c.JSON(tils)
}

// For create function use a JWT cookie with user_id like in the middleware.
// Extract user_id and verify a user with this user_id exists before
// adding it to TIL. The middleware stores the userID in c.Locals
func (h *TilHandler) Create(c *fiber.Ctx) error {
	var t til.TIL
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	exists, err := h.userService.UserExists(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	t.UserID = userID
	if err := t.Validate(); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.service.Create(t); err != nil {
		if errors.Is(err, til.ErrValidation) {
			return c.Status(422).JSON(fiber.Map{"error": "Validation failed", "details": err.Error()})
		}
		if errors.Is(err, til.ErrDuplicate) {
			return c.Status(409).JSON(fiber.Map{"error": "Duplicate entry"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error", "details": err.Error()})
	}
	return c.SendStatus(201)
}

func (h *TilHandler) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	til, err := h.service.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "TIL not found"})
	}

	return c.JSON(til)
}

func (h *TilHandler) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var til til.TIL
	if err := c.BodyParser(&til); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	til.ID = uint(id) // ensure ID matches URL

	updatedTIL, err := h.service.Update(til)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(updatedTIL)
}

func (h *TilHandler) Search(c *fiber.Ctx) error {
	type SearchRequest struct {
		Title    string `json:"title"`
		Category string `json:"category"`
	}

	var req SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	tils, err := h.service.Search(req.Title, req.Category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(tils)
}
