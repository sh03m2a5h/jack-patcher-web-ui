package groups

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AlsaDeviceIfInfo struct {
	Bits     []string `json:"bits"`
	Rate     int      `json:"rate"`
	Channels int      `json:"channels"`
}

type AlsaDeviceInfo struct {
	DeviceName string           `json:"device_name"`
	Playback   AlsaDeviceIfInfo `json:"playback"`
	Capture    AlsaDeviceIfInfo `json:"capture"`
}

func getAlsaDeviceSubDeviceCounts() map[string]int {
	output, err := exec.Command("aplay", "-l").CombinedOutput()
	if err != nil {
		log.Println(err)
		return nil
	}

	re := regexp.MustCompile(`card (\d+):`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	if matches == nil {
		return nil
	}

	deviceSubDeviceCounts := make(map[string]int)
	for _, match := range matches {
		deviceSubDeviceCounts[match[1]] = 0
	}

	return deviceSubDeviceCounts
}

func parseAlsaInfos(output string) (AlsaDeviceIfInfo, error) {
	ifInfo := AlsaDeviceIfInfo{}

	// -------------------- に囲まれた部分を取得
	re := regexp.MustCompile(`(?s)--------------------\n(.*?)\n--------------------`)
	allBracketsRe := regexp.MustCompile(`[\[\]\(\)\{\}]`)
	lastDigitRe := regexp.MustCompile(`\d+$`)

	matches := re.FindAllStringSubmatch(string(output), -1)
	if matches == nil {
		return ifInfo, errors.New("failed to get device info")
	}
	for _, match := range matches {
		// log.Println(match[1])
		matchStr := match[1]
		lines := strings.Split(matchStr, "\n")
		for _, line := range lines {
			kv := strings.Split(line, ":")
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			switch key {
			case "FORMAT":
				ifInfo.Bits = strings.Split(value, " ")
			case "CHANNELS":
				value = allBracketsRe.ReplaceAllString(value, "")
				ifInfo.Channels, _ = strconv.Atoi(lastDigitRe.FindString(value))
			case "RATE":
				value = allBracketsRe.ReplaceAllString(value, "")
				ifInfo.Rate, _ = strconv.Atoi(lastDigitRe.FindString(value))
			}

		}
	}

	return ifInfo, nil
}

func getAlsaDeviceInfo(deviceId string) (AlsaDeviceInfo, error) {
	// cat /proc/asound/card{N}/stream{M}
	deviceInfo := AlsaDeviceInfo{
		DeviceName: deviceId,
	}

	// command with timeout
	// playbackRes, _ := exec.Command("aplay", "-D", deviceId, "--dump-hw-params", "/dev/zero").CombinedOutput()
	playbackRes, _ := exec.Command("timeout", "0.5", "aplay", "-D", deviceId, "--dump-hw-params", "/dev/zero").CombinedOutput()
	ifInfo, err := parseAlsaInfos(string(playbackRes))
	if err != nil {
		return deviceInfo, err
	}
	deviceInfo.Playback = ifInfo

	captureRes, _ := exec.Command("timeout", "0.5", "arecord", "-D", deviceId, "--dump-hw-params", "/dev/null").CombinedOutput()
	ifInfo, err = parseAlsaInfos(string(captureRes))
	if err != nil {
		return deviceInfo, err
	}
	deviceInfo.Capture = ifInfo

	return deviceInfo, nil
}

func GetAlsaDevices(c *fiber.Ctx) error {
	// aplay -l
	subDeviceInfo := getAlsaDeviceSubDeviceCounts()

	devices := make([]AlsaDeviceInfo, 0, len(subDeviceInfo))
	for deviceName, subDeviceNum := range subDeviceInfo {
		deviceId := "hw:" + deviceName + "," + strconv.Itoa(subDeviceNum)
		log.Println(deviceId)
		deviceInfo, err := getAlsaDeviceInfo(deviceId)
		log.Println(deviceInfo)
		if err != nil {
			continue
		}
		devices = append(devices, deviceInfo)
	}

	return c.JSON(devices)
}
