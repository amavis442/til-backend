package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/config"
	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/middleware"
	"github.com/amavis442/til-backend/internal/til"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func waitForDB(dsn string, maxRetries int, delay time.Duration) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			if pingErr := sqlDB.Ping(); pingErr == nil {
				slog.Info("Connected to the database.")
				return db
			}
		}

		slog.Info(fmt.Sprintf("Waiting for database... (%d/%d)", i+1, maxRetries))
		time.Sleep(delay)
	}

	log.Fatalf("Could not connect to the database: %v", err)
	return nil
}

func main() {
	cfg := config.Load()
	if err := auth.InitJWTKeys(""); err != nil {
		log.Fatalf("failed to initialize JWT keys: %v", err)
	}

	dsn := cfg.DB_DSN
	port := cfg.PORT
	corsAllowedOrigin := cfg.CORSAllowedOrigin

	db := waitForDB(dsn, 10, 2*time.Second)
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// User and Auth
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	refreshTokenRepo := auth.NewRepository(db)
	refreshTokenService := auth.NewService(refreshTokenRepo)
	authHandler := handler.NewAuthHandler(userService, refreshTokenService, slogger)

	// Today I Learned (TIL)
	tilRepo := til.NewRepository(db)
	tilService := til.NewService(tilRepo)
	tilHandler := handler.NewTilHandler(tilService, userService)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsAllowedOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET,POST,OPTIONS,PUT",
		AllowCredentials: true,
	}))

	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "Europe/Amsterdam",
	}))

	auth := app.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)

	api := app.Group("/api", middleware.AuthMiddleware)
	api.Get("/tils", tilHandler.List)
	api.Post("/tils/search", tilHandler.Search)
	api.Get("/tils/:id", tilHandler.GetByID)
	api.Post("/tils", tilHandler.Create)
	api.Put("/tils/:id", tilHandler.Update)

	log.Fatal(app.Listen(":" + port))
}
