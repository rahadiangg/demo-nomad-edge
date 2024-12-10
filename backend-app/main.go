package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	appVersion = "v0.0.1" // default value, overridden by -ldflags
)

func main() {

	// Define CLI flag
	showVersion := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	// Check if the `--version` flag is set
	if *showVersion {
		fmt.Printf("%s", appVersion)
		os.Exit(0)
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	db, err := dbConn()
	if err != nil {
		slog.Error(fmt.Sprintf("failed connect database: %s", err.Error()))
	}

	validate := validator.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/transaction", func(c *fiber.Ctx) error {

		input := requestData{}

		if err := c.BodyParser(&input); err != nil {
			slog.Error(fmt.Sprintf("failed to parse: %s", err.Error()))
			return c.Status(fiber.StatusBadRequest).JSON(responseFormat{
				Message: "bad request",
			})
		}

		if err := validate.Struct(input); err != nil {
			slog.Error(fmt.Sprintf("failed to parse: %s", err.Error()))
			return c.Status(fiber.StatusBadRequest).JSON(responseFormat{
				Message: "bad request",
			})
		}

		dataDb := transaction{
			Platform: input.Platform,
			Amount:   input.Amount,
		}

		if err := db.WithContext(c.Context()).Create(&dataDb).Error; err != nil {
			slog.Error(fmt.Sprintf("failed save record data: %s", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(responseFormat{
				Message: "internal error",
			})
		}

		return c.JSON(responseFormat{
			Message: "OK",
		})
	})

	signalExit := make(chan os.Signal, 1)
	signal.Notify(signalExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	slog.Info("Server started at :8080")
	go func() {
		if err := app.Listen(":8080"); err != nil {
			slog.Error(fmt.Sprintf("failed start app: %s", err.Error()))
			os.Exit(1)
		}
	}()

	<-signalExit
	slog.Info("Got shutdown signal, try to shutdown gracefully")

	ctxDeadline, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctxDeadline); err != nil {
		slog.Error(fmt.Sprintf("can't shutdown gracefully: %s", err.Error()))
		os.Exit(1)
	}
	slog.Info("app shutdown gracefully")
}

func dbConn() (*gorm.DB, error) {

	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	db_port := os.Getenv("DB_PORT")

	dsn := "host=" + db_host + " user=" + db_user + " password=" + db_pass + " dbname=" + db_name + " port=" + db_port + " sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn))

	db.AutoMigrate(&transaction{})

	return db, err
}
