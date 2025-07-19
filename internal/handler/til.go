package handler

import (
	"strconv"

	"github.com/amavis442/til-backend/internal/til"
	"github.com/gofiber/fiber/v2"
)

type TilHandler struct {
	service til.Service
}

func NewTilHandler(s til.Service) *TilHandler {
	return &TilHandler{
		service: s,
	}
}

func (h *TilHandler) List(c *fiber.Ctx) error {
	tils, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch TILs"})
	}
	return c.JSON(tils)
}

func (h *TilHandler) Create(c *fiber.Ctx) error {
	var t til.TIL
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := h.service.Create(t); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save TIL"})
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
