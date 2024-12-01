package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"sh1pc.dev/jack-patcher-api/groups"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	app := fiber.New()

	groups.Initialize()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	jackGroup := app.Group("/api/jack")
	jackGroup.Get("/control/server", groups.GetJackServerState)
	jackGroup.Post("/control/server", groups.StartJackServer)
	jackGroup.Delete("/control/server", groups.StopJackServer)
	jackGroup.Get("/ports", groups.GetPorts)
	jackGroup.Get("/patches", groups.GetPatches)
	jackGroup.Post("/patches", groups.ConnectPorts)
	jackGroup.Delete("/patches", groups.DisconnectPorts)

	alsaGroup := app.Group("/api/alsa")
	alsaGroup.Get("/devices", groups.GetAlsaDevices)
	alsaGroup.Post("/load", groups.LoadAlsaDeviceToJack)
	alsaGroup.Delete("/load/:device_id", groups.UnloadAlsaDeviceFromJack)

	log.Fatal(app.Listen(":3000"))
}
