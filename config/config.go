// Package config handles reading and parsing the application's configuration,
// primarily the list of devices to be backed up.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/models"
)

// ReadConfig reads the device inventory from the specified JSON file.
func ReadConfig() ([]models.Device, error) {
	// Use os.ReadFile, it's the modern standard
	data, err := os.ReadFile("devices/devices.json")
	if err != nil {
		return nil, fmt.Errorf("error reading devices.json: %w", err)
	}

	var devices []models.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("error parsing devices.json: %w", err)
	}

	return devices, nil
}
