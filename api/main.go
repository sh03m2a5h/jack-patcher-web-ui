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
		// devices, _ := groups.GetAlsaDeviceInfo("hw:0,0")
		// return c.JSON(devices)
	})
	jackGroup := app.Group("/api/jack/control")
	jackGroup.Get("/server", groups.GetJackServerState)
	jackGroup.Post("/server", groups.StartJackServer)
	jackGroup.Delete("/server", groups.StopJackServer)

	alsaGroup := app.Group("/api/alsa")
	alsaGroup.Get("/devices", groups.GetAlsaDevices)
	alsaGroup.Post("/load", groups.LoadAlsaDeviceToJack)
	alsaGroup.Delete("/load/:device_id", groups.UnloadAlsaDeviceFromJack)

	log.Fatal(app.Listen(":3000"))
}
