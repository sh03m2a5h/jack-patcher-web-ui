package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"sh1pc.dev/jack-patcher-api/groups"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	jackGroup := app.Group("/api/jack")
	jackGroup.Get("/control/server", groups.GetJackServerState)
	jackGroup.Post("/control/server", groups.StartJackServer)
	jackGroup.Delete("/control/server", groups.StopJackServer)

	log.Fatal(app.Listen(":3000"))
}
