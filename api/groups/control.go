package groups

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type JackServerConfig struct {
	Rate     int `json:"rate"`
	Period   int `json:"period"`
	// dummyの場合はnperiodsを設定できない
	// NPeriods int `json:"nperiods"`
}

func GetJackServerState(c *fiber.Ctx) error {
	output, err := exec.Command("jack_control", "status").CombinedOutput()
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get JACK server status"})
	}
	return c.JSON(fiber.Map{"status": string(output)})
}

func StartJackServer(c *fiber.Ctx) error {
	var config JackServerConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Start JACK server
	if err := exec.Command("jack_control", "start").Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "JACK server already started"})
	}

	// Set device
	if err := exec.Command("jack_control", "ds", "dummy").Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to set driver"})
	}

	// Set parameters
	if err := exec.Command("jack_control", "dps", "rate", strconv.Itoa(config.Rate)).Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to set rate"})
	}
	if err := exec.Command("jack_control", "dps", "period", strconv.Itoa(config.Period)).Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to set period"})
	}

	return c.JSON(fiber.Map{"message": "JACK server started"})
}

func StopJackServer(c *fiber.Ctx) error {
	if err := exec.Command("jack_control", "stop").Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "JACK server already stopped"})
	}
	if err := exec.Command("jack_control", "exit").Run(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to exit JACK server"})
	}
	return c.JSON(fiber.Map{"message": "JACK server stopped"})
}
