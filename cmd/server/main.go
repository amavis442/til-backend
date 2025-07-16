package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/amavis442/til-backend/config"
	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/repository"
	"github.com/amavis442/til-backend/internal/usecase"
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
				log.Println("Connected to the database.")
				return db
			}
		}

		log.Printf("Waiting for database... (%d/%d)", i+1, maxRetries)
		time.Sleep(delay)
	}

	log.Fatalf("Could not connect to the database: %v", err)
	return nil
}

func main() {
	config.LoadEnv()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set")
	}

	db := waitForDB(dsn, 10, 2*time.Second)

	port := fmt.Sprint(os.Getenv("PORT"))

	tilRepo := repository.NewTILRepository(db)
	tilUsecase := usecase.NewTILUsecase(tilRepo)
	tilHandler := handler.NewTILHandler(tilUsecase)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Or "http://localhost:5173" if you want to restrict
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "Europe/Amsterdam",
	}))

	api := app.Group("/api")
	api.Get("/tils", tilHandler.List)
	api.Get("/tils/search", tilHandler.Search)
	api.Get("/tils/:id", tilHandler.GetByID)
	api.Post("/tils", tilHandler.Create)
	api.Put("/tils/:id", tilHandler.Update)

	log.Fatal(app.Listen(":" + port))
}
