package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/joho/godotenv"
	"github.com/mertvasit/go-url-shortener/routes"
	"log"
	"os"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("[ERROR] - Cannot load environment variables - ", err)
	}

	app := fiber.New()
	app.Use(middleware.Logger())

	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
