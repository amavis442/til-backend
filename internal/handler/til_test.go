package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/middleware"
	"github.com/amavis442/til-backend/internal/til"
	"github.com/amavis442/til-backend/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestApp(verifier auth.TokenVerifier) (*fiber.App, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file:testdb?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&user.User{}, &til.TIL{})
	db.Create(&user.User{
		Model:        gorm.Model{ID: 1},
		Username:     "testuser",
		PasswordHash: "irrelevant",
	})

	userRepo := user.NewRepository(db)
	repo := til.NewRepository(db)
	uc := til.NewService(repo)
	userService := user.NewService(userRepo)
	h := handler.NewTilHandler(uc, userService)

	app := fiber.New()
	api := app.Group("/api", middleware.AuthMiddleware(verifier))
	api.Get("/tils", h.List)
	api.Post("/tils", h.Create)

	return app, db
}

func TestCreateAndListTIL(t *testing.T) {
	verifier := &mockTokenVerifier{}
	app, _ := setupTestApp(verifier)

	// Step 1: Create a TIL via POST
	input := til.TIL{
		Title:    "TIL Go tests are fun",
		Content:  "Writing tests with Fiber and GORM",
		Category: "golang",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/tils", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer dummy-token") // required for middleware
	resp, err := app.Test(req)

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Response body: %s", string(bodyBytes))
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Step 2: Get the list of TILs
	req = httptest.NewRequest(http.MethodGet, "/api/tils", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer dummy-token") // required again
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tils []til.TIL
	var respBody handler.Response[[]til.TIL]
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	tils = respBody.Items

	assert.Len(t, tils, 1)
	if len(tils) > 0 {
		assert.Len(t, tils, 1)
		assert.Equal(t, input.Title, tils[0].Title)
		assert.Equal(t, input.Content, tils[0].Content)
		assert.Equal(t, input.Category, tils[0].Category)
		assert.WithinDuration(t, time.Now(), tils[0].CreatedAt, time.Second)
	}
}
