package handler

import (
	"strconv"

	"github.com/amavis442/til-backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type TILHandler struct {
	usecase domain.TILUsecase
}

func NewTILHandler(u domain.TILUsecase) *TILHandler {
	return &TILHandler{u}
}

func (h *TILHandler) List(c *fiber.Ctx) error {
	tils, err := h.usecase.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch TILs"})
	}
	return c.JSON(tils)
}

func (h *TILHandler) Create(c *fiber.Ctx) error {
	var t domain.TIL
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := h.usecase.Create(t); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save TIL"})
	}
	return c.SendStatus(201)
}

func (h *TILHandler) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	til, err := h.usecase.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "TIL not found"})
	}

	return c.JSON(til)
}

func (h *TILHandler) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var til domain.TIL
	if err := c.BodyParser(&til); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	til.ID = uint(id) // ensure ID matches URL

	updatedTIL, err := h.usecase.Update(til)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(updatedTIL)
}

func (h *TILHandler) Search(c *fiber.Ctx) error {
	title := c.Query("title")
	category := c.Query("category")

	tils, err := h.usecase.Search(title, category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(tils)
}
