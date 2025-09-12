// config.go
package config

import (
	"encoding/json"
	"log"
	"os"
	"ssh-fetcher/models"
)

func ReadConfig() []models.Device {
	// Используем os.ReadFile, это современный стандарт
	data, err := os.ReadFile("devices/devices.json") 
	if err != nil {
		log.Fatalf("Ошибка чтения devices.json: %v", err)
	}

	var devices []models.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		log.Fatalf("Ошибка парсинга devices.json: %v", err)
	}

	return devices
}