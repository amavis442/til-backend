package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/til"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestApp() (*fiber.App, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&til.TIL{})

	repo := til.NewRepository(db)
	uc := til.NewService(repo)
	h := handler.NewTilHandler(uc)

	app := fiber.New()
	api := app.Group("/api")
	api.Get("/tils", h.List)
	api.Post("/tils", h.Create)

	return app, db
}

func TestCreateAndListTIL(t *testing.T) {
	app, _ := setupTestApp()

	// Step 1: Create a TIL via POST
	input := til.TIL{
		Title:    "TIL Go tests are fun",
		Content:  "Writing tests with Fiber and GORM",
		Category: "golang",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/tils", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Step 2: Get the list of TILs
	req = httptest.NewRequest(http.MethodGet, "/api/tils", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tils []til.TIL
	json.NewDecoder(resp.Body).Decode(&tils)

	assert.Len(t, tils, 1)
	assert.Equal(t, input.Title, tils[0].Title)
	assert.Equal(t, input.Content, tils[0].Content)
	assert.Equal(t, input.Category, tils[0].Category)
	assert.WithinDuration(t, time.Now(), tils[0].CreatedAt, time.Second)
}
