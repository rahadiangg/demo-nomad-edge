package main

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CustomHandler struct {
	db     *gorm.DB
	config CurrentConfig
}

func NewHandler(db *gorm.DB, config CurrentConfig) *CustomHandler {
	return &CustomHandler{
		db:     db,
		config: config,
	}
}

func (h *CustomHandler) LocalDashboard(c *fiber.Ctx) error {

	var totalData, dataSent, dataUnsent int64

	if err := h.db.WithContext(c.Context()).Model(&localTransaction{}).Count(&totalData).Error; err != nil {
		slog.Error(fmt.Sprintf("failed count all transation data: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).SendString("internal error")
	}

	if err := h.db.WithContext(c.Context()).Model(&localTransaction{}).Where("send_at IS NOT NULL").Count(&dataSent).Error; err != nil {
		slog.Error(fmt.Sprintf("failed count data sent: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).SendString("internal error")
	}

	if err := h.db.WithContext(c.Context()).Model(&localTransaction{}).Where("send_at IS NULL").Count(&dataUnsent).Error; err != nil {
		slog.Error(fmt.Sprintf("failed count data unsent: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).SendString("internal error")
	}

	return c.Render("localDashboard", fiber.Map{
		"appVersion":               appVersion,
		"configDatabasePath":       h.config.DatabasePath,
		"configIntervalRandomData": h.config.IntervalRandomData,
		"configIntervalSendData":   h.config.IntervalSendData,
		"configPlatform":           h.config.Platform,
		"configBackendUri":         h.config.BackendUri,
		"totalData":                totalData,
		"dataSend":                 dataSent,
		"dataUnsend":               dataUnsent,
	})
}
