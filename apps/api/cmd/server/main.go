package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status": "ok",
			},
		})
	})

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
