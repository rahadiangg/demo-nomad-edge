package main

import (
	"gorm.io/gorm"
)

type requestData struct {
	Platform string `json:"platform" validate:"required"`
	Amount   uint32 `json:"amount" validate:"required"`
}

type responseFormat struct {
	Message string `json:"message"`
}

type transaction struct {
	*gorm.Model
	Platform string
	Amount   uint32
}
