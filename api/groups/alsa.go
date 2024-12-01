package groups

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	_ "github.com/mattn/go-sqlite3"
)

type AlsaDevice struct {
	Id         string
	CardName   string
	DeviceName string
	CardNum    int
	DeviceNum  int
}

type AlsaDeviceIfInfo struct {
	BitFormats  []string // e.g. ["S16_LE", "S32_LE"]
	MinRate     int
	MaxRate     int
	MaxChannels int
	MinPeriod   int
	MaxPeriod   int
}

type DevicePadInfo struct {
	Rate     int `json:"rate"`
	Channels int `json:"channels"`
}

type AlsaDeviceInfo struct {
	Id         string         `json:"id"`
	CardName   string         `json:"card_name"`
	DeviceName string         `json:"device_name"`
	Playback   *DevicePadInfo `json:"playback"`
	Capture    *DevicePadInfo `json:"capture"`
}

type AlsaDeviceConnectionType string

const (
	zAlsa AlsaDeviceConnectionType = "zalsa"
	alsa  AlsaDeviceConnectionType = "alsa"
)

type AlsaLoadConfig struct {
	DeviceId string                   `json:"deviceId"`
	Client   AlsaDeviceConnectionType `json:"client"`
	Rate     int                      `json:"rate"`
	// CaptureChannels  int    `json:"capture_channels"`
	// PlaybackChannels int    `json:"playback_channels"`
	// Format           string `json:"format"`
	Period   int `json:"period"`
	NPeriods int `json:"nperiods"`
}

func openDb() (*sql.DB, error) {
	return sql.Open("sqlite3", "./audio_devices.db")
}

func InitializeDb() {
	// SQLite3 データベースを開く
	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// デバイス情報を保存するテーブルを初期化
	initializeDevicesTable(db)

	// 復帰用テーブルを初期化
	initializeLoadDevicesTable(db)
	initializePatchesTable(db)
}

func InitializeDevices() {
	log.Println("Initializing alsa devices")
	defer log.Println("Finished initializing alsa devices")

	// SQLite3 データベースを開く
	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// デバイス情報を更新
	updateDecicesTable(db)

	// 保存されたデバイスをロード
	loadSavedDevices(db)

	// 保存されたパッチをロード
	restorePortConnections(db)
}

func initializeDevicesTable(db *sql.DB) {
	// テーブルが存在しない場合は作成
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT,
		card_name TEXT,
		device_name TEXT,
		card_number INT,
		device_number INT,
		PRIMARY KEY (id)
	);
	CREATE TABLE IF NOT EXISTS device_c_params (
		device_id TEXT,
		min_bits INT,
		max_bits INT,
		min_rate INT,
		max_rate INT,
		max_channels INT,
		min_period INT,
		max_period INT,
		PRIMARY KEY (device_id)
		FOREIGN KEY (device_id) REFERENCES devices(id)
	);
	CREATE TABLE IF NOT EXISTS device_p_params (
		device_id TEXT,
		min_bits INT,
		max_bits INT,
		min_rate INT,
		max_rate INT,
		max_channels INT,
		min_period INT,
		max_period INT,
		PRIMARY KEY (device_id)
		FOREIGN KEY (device_id) REFERENCES devices(id)
	);
	`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func updateDecicesTable(db *sql.DB) {
	// aplay -l コマンドでデバイス一覧を取得
	cmd := exec.Command("aplay", "-l")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	cardDeviceRegex := regexp.MustCompile(`card (\d+): (.+), device (\d+): (.+)`)
	matches := cardDeviceRegex.FindAllStringSubmatch(string(output), -1)
	if matches == nil {
		log.Fatal("failed to parse aplay output")
	}

	// カード名とデバイス名の対応をデータベースに保存
	for _, match := range matches {
		cardIndex, _ := strconv.Atoi(match[1])
		cardName := match[2]
		deviceIndex, _ := strconv.Atoi(match[3])
		deviceName := match[4]
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cardName+deviceName)))[0:16]

		alsaDevice := AlsaDevice{
			Id:         hash,
			CardName:   cardName,
			DeviceName: deviceName,
			CardNum:    cardIndex,
			DeviceNum:  deviceIndex,
		}

		_, err = db.Exec(`INSERT INTO devices (id, card_name, device_name, card_number, device_number) 
			VALUES (?, ?, ?, ?, ?) 
			ON CONFLICT(id) 
			DO UPDATE SET card_name=excluded.card_name, device_name=excluded.device_name, card_number=excluded.card_number, device_number=excluded.device_number`,
			alsaDevice.Id, alsaDevice.CardName, alsaDevice.DeviceName, alsaDevice.CardNum, alsaDevice.DeviceNum)
		if err != nil {
			log.Fatal(err)
		}

		hwId := fmt.Sprintf("hw:%d,%d", cardIndex, deviceIndex)
		playbackParams, err := getDeviceParams(exec.Command("timeout", "0.5", "aplay", "--dump-hw-params", "-D", hwId, "/dev/zero"))
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = db.Exec(`INSERT INTO device_p_params (device_id, min_bits, max_bits, min_rate, max_rate, max_channels, min_period, max_period)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(device_id)
			DO UPDATE SET min_bits=excluded.min_bits, max_bits=excluded.max_bits, min_rate=excluded.min_rate, max_rate=excluded.max_rate, max_channels=excluded.max_channels, min_period=excluded.min_period, max_period=excluded.max_period`,
			alsaDevice.Id, 16, 32, playbackParams.MinRate, playbackParams.MaxRate, playbackParams.MaxChannels, playbackParams.MinPeriod, playbackParams.MaxPeriod)
		if err != nil {
			log.Fatal(err)
		}

		captureParams, err := getDeviceParams(exec.Command("timeout", "0.5", "arecord", "--dump-hw-params", "-D", hwId, "/dev/null"))
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = db.Exec(`INSERT INTO device_c_params (device_id, min_bits, max_bits, min_rate, max_rate, max_channels, min_period, max_period) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(device_id)
			DO UPDATE SET min_bits=excluded.min_bits, max_bits=excluded.max_bits, min_rate=excluded.min_rate, max_rate=excluded.max_rate, max_channels=excluded.max_channels, min_period=excluded.min_period, max_period=excluded.max_period`,
			alsaDevice.Id, 16, 32, captureParams.MinRate, captureParams.MaxRate, captureParams.MaxChannels, captureParams.MinPeriod, captureParams.MaxPeriod)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func initializeLoadDevicesTable(db *sql.DB) {
	createLoadedDevicesTableSQL := `
	CREATE TABLE IF NOT EXISTS loaded_devices (
		device_id TEXT PRIMARY KEY,
		client_type TEXT,
		rate INT,
		period INT,
		nperiods INT,
		FOREIGN KEY (device_id) REFERENCES devices(id)
	);
	`
	_, err := db.Exec(createLoadedDevicesTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func loadSavedDevices(db *sql.DB) {
	rows, err := db.Query(`SELECT device_id, client_type, rate, period, nperiods FROM loaded_devices`)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var config AlsaLoadConfig
		err := rows.Scan(&config.DeviceId, &config.Client, &config.Rate, &config.Period, &config.NPeriods)
		if err != nil {
			log.Println(err)
			continue
		}
		// デバイスを再ロード
		err = loadAlsaDevice(config)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func getDeviceParams(cmd *exec.Cmd) (AlsaDeviceIfInfo, error) {
	ifInfo := AlsaDeviceIfInfo{}

	// cmd := exec.Command("timeout", "0.5", "aplay", "--dump-hw-params", "-D", deviceIdentifier, "/dev/zero")
	output, _ := cmd.CombinedOutput()

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
				ifInfo.BitFormats = strings.Split(value, " ")
			case "CHANNELS":
				value = allBracketsRe.ReplaceAllString(value, "")
				ifInfo.MaxChannels, _ = strconv.Atoi(lastDigitRe.FindString(value))
			case "RATE":
				value = allBracketsRe.ReplaceAllString(value, "")
				splitted := strings.Split(value, " ")
				ifInfo.MinRate, _ = strconv.Atoi(splitted[0])
				if len(splitted) > 1 {
					ifInfo.MaxRate, _ = strconv.Atoi(splitted[1])
				} else {
					ifInfo.MaxRate = ifInfo.MinRate
				}
			case "PERIOD":
				value = allBracketsRe.ReplaceAllString(value, "")
				splitted := strings.Split(value, " ")
				ifInfo.MinPeriod, _ = strconv.Atoi(splitted[0])
				ifInfo.MaxPeriod, _ = strconv.Atoi(splitted[1])
			}

		}
	}

	return ifInfo, nil
}

func GetAlsaDevices(c *fiber.Ctx) error {
	devices := make([]AlsaDeviceInfo, 0)

	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, card_name, device_name, card_number, device_number
		FROM devices
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var device AlsaDeviceInfo
		var deviceId, cardName, deviceName string
		var cHwNumber, dHwNumber int

		err := rows.Scan(&deviceId, &cardName, &deviceName, &cHwNumber, &dHwNumber)
		if err != nil {
			return err
		}

		device.Id = deviceId
		device.CardName = cardName
		device.DeviceName = deviceName

		playbackParams := db.QueryRow(`
			SELECT min_rate, max_rate, max_channels
			FROM device_p_params
			WHERE device_id = ?
		`, deviceId)

		var minRate, maxRate, maxChannels int
		err = playbackParams.Scan(&minRate, &maxRate, &maxChannels)
		if err != nil {
			log.Println(err)
			// return err
		} else {
			device.Playback = &DevicePadInfo{
				Rate:     maxRate,
				Channels: maxChannels,
			}
		}

		captureParams := db.QueryRow(`
			SELECT min_rate, max_rate, max_channels
			FROM device_c_params
			WHERE device_id = ?
		`, deviceId)

		err = captureParams.Scan(&minRate, &maxRate, &maxChannels)
		if err != nil {
			log.Println(err)
			// return err
		} else {
			device.Capture = &DevicePadInfo{
				Rate:     maxRate,
				Channels: maxChannels,
			}
		}

		devices = append(devices, device)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return c.JSON(devices)
}

func LoadAlsaDeviceToJack(c *fiber.Ctx) error {
	// deviceName := c.Params("device_name")
	var config AlsaLoadConfig

	err := c.BodyParser(&config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	alsaDevice := AlsaDevice{}

	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}

	err = db.QueryRow(`
		SELECT id, card_name, device_name, card_number, device_number
		FROM devices
		WHERE id = ?
	`, config.DeviceId).Scan(&alsaDevice.Id, &alsaDevice.CardName, &alsaDevice.DeviceName, &alsaDevice.CardNum, &alsaDevice.DeviceNum)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get ALSA device info"})
	}

	// デバイスをロードする処理
	err = loadAlsaDevice(config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to load ALSA device to JACK"})
	}

	// デバイス情報をデータベースに保存
	_, err = db.Exec(`INSERT INTO loaded_devices (device_id, client_type, rate, period, nperiods) VALUES (?, ?, ?, ?, ?)`,
		config.DeviceId, config.Client, config.Rate, config.Period, config.NPeriods)
	if err != nil {
		log.Println(err)
	}

	return c.JSON(fiber.Map{"message": "ALSA device loaded to JACK"})
}

// デバイスをロードする関数を追加
func loadAlsaDevice(config AlsaLoadConfig) error {
	alsaDevice := AlsaDevice{}

	db, err := openDb()
	if err != nil {
		log.Fatal(err)
	}

	err = db.QueryRow(`
		SELECT id, card_name, device_name, card_number, device_number
		FROM devices
		WHERE id = ?
	`, config.DeviceId).Scan(&alsaDevice.Id, &alsaDevice.CardName, &alsaDevice.DeviceName, &alsaDevice.CardNum, &alsaDevice.DeviceNum)
	if err != nil {
		return err
	}

	hwId := fmt.Sprintf("hw:%d,%d", alsaDevice.CardNum, alsaDevice.DeviceNum)

	var deviceConnectionType AlsaDeviceConnectionType
	if config.Client == "zalsa" {
		deviceConnectionType = zAlsa
	} else {
		deviceConnectionType = alsa
	}

	if deviceConnectionType == zAlsa {
		zalsaArg := fmt.Sprintf("-d %s -r %d -p %d -n %d", hwId, config.Rate, config.Period, config.NPeriods)

		if err := exec.Command("jack_load", alsaDevice.Id+"_in", "zalsa_in", "-i", zalsaArg).Run(); err != nil {
			log.Println(err)
			return err
		}

		if err := exec.Command("jack_load", alsaDevice.Id+"_out", "zalsa_out", "-i", zalsaArg).Run(); err != nil {
			log.Println(err)
			return err
		}
	} else if deviceConnectionType == alsa {
		if err := exec.Command("alsa_in", "-j", alsaDevice.Id+"_in", "-d", hwId, "-r", strconv.Itoa(config.Rate), "-p", strconv.Itoa(config.Period), "-n", strconv.Itoa(config.NPeriods)).Run(); err != nil {
			log.Println(err)
			return err
		}
		if err := exec.Command("alsa_out", "-j", alsaDevice.Id+"_out", "-d", hwId, "-r", strconv.Itoa(config.Rate), "-p", strconv.Itoa(config.Period), "-n", strconv.Itoa(config.NPeriods)).Run(); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func UnloadAlsaDeviceFromJack(c *fiber.Ctx) error {
	deviceId := c.Params("device_id")

	err := exec.Command("jack_unload", deviceId+"_in").Run()
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to unload ALSA device from JACK"})
	}
	err = exec.Command("jack_unload", deviceId+"_out").Run()
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to unload ALSA device from JACK"})
	}

	// データベースからデバイス情報を削除
	db, err := openDb()
	if err != nil {
		log.Println(err)
	}
	_, err = db.Exec(`DELETE FROM loaded_devices WHERE device_id = ?`, deviceId)
	if err != nil {
		log.Println(err)
	}

	return c.JSON(fiber.Map{"message": "ALSA device unloaded from JACK"})
}
