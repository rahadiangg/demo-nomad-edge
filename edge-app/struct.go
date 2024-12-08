package main

import (
	"time"

	"gorm.io/gorm"
)

type localTransaction struct {
	*gorm.Model
	Platform string
	Amount   uint32
	SendAt   *time.Time //nullable
}

type CurrentConfig struct {
	DatabasePath       string
	IntervalRandomData uint8
	IntervalSendData   uint8
	Platform           string
	BackendUri         string
}

type backendFormatRequest struct {
	Platform string `json:"platform"`
	Amount   uint32 `json:"amount"`
}
