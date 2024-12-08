package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"math/rand"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed *.html
var viewsfs embed.FS

func main() {

	// load configudation
	appIntervalRandomData, _ := strconv.Atoi(getEnv("APP_INTERVAL_RANDOM_DATA", "5"))
	appIntervalSendData, _ := strconv.Atoi(getEnv("APP_INTERVAL_SEND_DATA", "10"))

	config := CurrentConfig{
		DatabasePath:       getEnv("DB_PATH", "local.db"),
		IntervalRandomData: uint8(appIntervalRandomData),
		IntervalSendData:   uint8(appIntervalSendData),
		Platform:           getEnv("PLATFORM", "local"),
		BackendUri:         getEnv("BACKEND_URI", "http://please-change"),
	}

	db, err := dbConn(config.DatabasePath)
	if err != nil {
		slog.Error(fmt.Sprintf("failed init database: %s", err.Error()))
		os.Exit(1)
	}

	handler := NewHandler(db, config)

	engine := html.NewFileSystem(http.FS(viewsfs), ".html")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
	})

	app.Get("/", handler.LocalDashboard)

	signalExit := make(chan os.Signal, 1)
	signal.Notify(signalExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := app.Listen(":8080"); err != nil {
			slog.Error(fmt.Sprintf("failed start web server: %s", err.Error()))
			os.Exit(1)
		}
	}()

	// Create a new random source and generator
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Define the range for the random price
	minPrice := 1000
	maxPrice := 15000

	var wgLocalDb sync.WaitGroup

	// loop store local db
	intervalRandomData := time.Duration(config.IntervalRandomData) * time.Second
	go func() {
		for {
			wgLocalDb.Add(1)
			price := random.Intn(maxPrice-minPrice+1) + minPrice

			slog.Info(fmt.Sprintf("Random Price: %d\n", price))

			data := localTransaction{
				Platform: config.Platform,
				Amount:   uint32(price),
			}
			if err := db.Create(&data).Error; err != nil {
				slog.Error(fmt.Sprintf("failed inserd to loca database: %s", err.Error()))
			} else {
				slog.Info("That price inserted to local database")
			}

			wgLocalDb.Done()
			time.Sleep(intervalRandomData)
		}
	}()

	var wgSendData sync.WaitGroup

	// loop send data
	intervalSentData := time.Duration(config.IntervalSendData) * time.Second
	go func() {
		for {
			wgSendData.Add(1)

			if err := sendDataToBackend(db, &config); err != nil {
				slog.Error(fmt.Sprintf("failed send data: %s", err.Error()))
			}

			wgSendData.Done()
			time.Sleep(intervalSentData)
		}
	}()

	<-signalExit
	slog.Info("Got shutdown signal, try to shutdown gracefully")

	wgLocalDb.Wait()
	wgSendData.Wait()
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Error(fmt.Sprintf("failed shutdown web server gracefully: %s", err.Error()))
	} else {
		slog.Info("App shutdown gracefully")
	}

}

func dbConn(dbPath string) (*gorm.DB, error) {

	db, err := gorm.Open(sqlite.Open(dbPath))

	db.AutoMigrate(&localTransaction{})

	return db, err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func sendDataToBackend(db *gorm.DB, config *CurrentConfig) error {

	var data localTransaction

	if err := db.Model(&localTransaction{}).First(&data, "send_at IS NULL").Error; err != nil {
		errMsg := fmt.Errorf("can't find unsent data:: %s", err.Error())
		return errMsg
	}

	payload := backendFormatRequest{
		Platform: data.Platform,
		Amount:   data.Amount,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		errMsg := fmt.Errorf("failed marshal backend payload: %s", err.Error())
		return errMsg
	}

	resp, err := http.Post(config.BackendUri, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		errMsg := fmt.Errorf("failed send data to backend: %s", err.Error())
		return errMsg
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("backend response non 200 OK")
	}

	// data send success, mark the sendAt
	sendAt := time.Now()
	data.SendAt = &sendAt

	if err := db.Save(&data).Error; err != nil {
		errMsg := fmt.Errorf("failed update send_at: %s", err.Error())
		return errMsg
	}

	return nil
}
