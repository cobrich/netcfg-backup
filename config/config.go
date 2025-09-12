package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"ssh-fetcher/models"
)

func ReadConfig() []models.Device {
	data, err := ioutil.ReadFile("devices/devices.json")
	if err != nil {
		log.Fatalf("Ошибка чтения devices.json: %v", err)
	}

	var devices []models.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		log.Fatalf("Ошибка парсинга devices.json: %v", err)
	}

	return devices
}
