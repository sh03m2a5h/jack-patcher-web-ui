package groups

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JackIoType string

const (
	In  JackIoType = "in"
	Out JackIoType = "out"
)

func parseIoType(ioType string) JackIoType {
	if strings.Contains(ioType, "in") {
		return In
	}
	return Out
}

type JackPort struct {
	IoType  JackIoType `json:"io_type"`
	Id      string     `json:"id"`
	Channel int        `json:"channel"`
}

type JackPatch struct {
	Source      JackPort `json:"source" binding:"required"`
	Destination JackPort `json:"destination" binding:"required"`
}

type JackPortConnection struct {
	Source       JackPort   `json:"source"`
	Destinations []JackPort `json:"destinations"`
}

func GetPorts(c *fiber.Ctx) error {
	out, err := exec.Command("jack_lsp").Output()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get ports"})
	}

	portRe := regexp.MustCompile(`^([a-z0-9]+)_(in:capture|out:playback)_(\d)+`)

	ports := make([]JackPort, 0)
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		match := portRe.FindStringSubmatch(line)
		if len(match) > 0 {
			id := match[1]
			ioType := parseIoType(match[2]) // in or out
			channel, _ := strconv.Atoi(match[3])

			ports = append(ports, JackPort{
				IoType:  ioType,
				Id:      id,
				Channel: channel,
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ports": ports})
}

func GetPatches(c *fiber.Ctx) error {
	out, err := exec.Command("jack_lsp", "-c").Output()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get ports"})
	}

	srcRe := regexp.MustCompile(`^([a-z0-9]+)_in:capture_(\d)+`)
	destRe := regexp.MustCompile(`^\s{1,4}([a-z0-9]+)_out:playback_(\d)+`)

	patches := make([]JackPortConnection, 0)
	lines := strings.Split(string(out), "\n")

	var source JackPort

	for _, line := range lines {
		srcMatch := srcRe.FindStringSubmatch(line)
		if len(srcMatch) > 0 {
			sourceId := srcMatch[1]
			channel, _ := strconv.Atoi(srcMatch[2])

			source = JackPort{
				IoType:  In,
				Id:      sourceId,
				Channel: channel,
			}
			patches = append(patches, JackPortConnection{
				Source:       source,
				Destinations: make([]JackPort, 0),
			})
			continue
		}

		destMatch := destRe.FindStringSubmatch(line)
		if len(destMatch) > 0 {
			destId := destMatch[1]
			channel, _ := strconv.Atoi(destMatch[2])

			dest := JackPort{
				IoType:  Out,
				Id:      destId,
				Channel: channel,
			}

			lastIndex := len(patches) - 1
			patches[lastIndex].Destinations = append(patches[lastIndex].Destinations, dest)
		}
	}

	// delete empty patches
	for i := 0; i < len(patches); i++ {
		if len(patches[i].Destinations) == 0 {
			patches = append(patches[:i], patches[i+1:]...)
			i--
		}
	}

	return c.Status(fiber.StatusOK).JSON(patches)
}

func ConnectPorts(c *fiber.Ctx) error {
	var patch JackPatch
	if err := c.BodyParser(&patch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	if patch.Source.IoType != In || patch.Destination.IoType != Out {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid port types"})
	}

	srcStr := patch.Source.Id + "_in:capture_" + strconv.Itoa(patch.Source.Channel)
	destStr := patch.Destination.Id + "_out:playback_" + strconv.Itoa(patch.Destination.Channel)

	err := exec.Command("jack_connect", srcStr, destStr).Run()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to connect ports"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Ports connected"})
}

func DisconnectPorts(c *fiber.Ctx) error {
	var patch JackPatch
	// print body
	println(string(c.Body()))

	if err := c.BodyParser(&patch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}

	if patch.Source.IoType != In || patch.Destination.IoType != Out {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid port types"})
	}

	srcStr := patch.Source.Id + "_in:capture_" + strconv.Itoa(patch.Source.Channel)
	destStr := patch.Destination.Id + "_out:playback_" + strconv.Itoa(patch.Destination.Channel)

	err := exec.Command("jack_disconnect", srcStr, destStr).Run()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to disconnect ports"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Ports disconnected"})
}
