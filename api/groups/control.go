package groups

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type JackServerConfig struct {
	Rate   int `json:"rate"`
	Period int `json:"period"`
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

func StartJack(config JackServerConfig) (string, error) {
	if err := exec.Command("jack_control", "start").Run(); err != nil {
		return "JACK server already started", nil
	}

	// Start JACK server
	if err := exec.Command("jack_control", "start").Run(); err != nil {
		log.Println(err)
		return "JACK server already started", nil
	}

	// Set device
	if err := exec.Command("jack_control", "ds", "dummy").Run(); err != nil {
		log.Println(err)
		return "Failed to set driver", nil
	}

	// Set parameters
	if err := exec.Command("jack_control", "dps", "rate", strconv.Itoa(config.Rate)).Run(); err != nil {
		log.Println(err)
		return "Failed to set rate", nil
	}
	if err := exec.Command("jack_control", "dps", "period", strconv.Itoa(config.Period)).Run(); err != nil {
		log.Println(err)
		return "Failed to set period", nil
	}

	// SQLite3 データベースを開く
	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}

	// 保存されたデバイスをロード
	loadSavedDevices(db)

	// 保存されたパッチをロード
	restorePortConnections(db)

	return "JACK server started", nil
}

func StartJackServer(c *fiber.Ctx) error {
	var config JackServerConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// JACKサーバーを起動
	message, err := StartJack(config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": message})
	}

	return c.JSON(fiber.Map{"message": message})
}

func StopJack() (string, error) {
	if err := exec.Command("jack_control", "stop").Run(); err != nil {
		return "JACK server already stopped", nil
	}
	if err := exec.Command("jack_control", "exit").Run(); err != nil {
		log.Println(err)
		return "Failed to exit JACK server", nil
	}
	return "JACK server stopped", nil
}

func StopJackServer(c *fiber.Ctx) error {
	message, err := StopJack()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": message})
	}
	return c.JSON(fiber.Map{"message": message})
}
