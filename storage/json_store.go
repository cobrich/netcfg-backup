package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/models"
)

// JSONStore implements the Store interface using a JSON file.
type JSONStore struct {
	filePath string
}

// NewJSONStore creates a new instance of a JSONStore.
func NewJSONStore(filePath string) *JSONStore {
	return &JSONStore{filePath: filePath}
}

// GetAllDevices reads and parses the device list from the JSON file.
func (js *JSONStore) GetAllDevices() ([]models.Device, error) {
	data, err := os.ReadFile(js.filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading devices file %s: %w", js.filePath, err)
	}

	var devices []models.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("error parsing devices file %s: %w", js.filePath, err)
	}

	return devices, nil
}

// AddDevice adds a new device to the JSON file.
// It reads the existing devices, appends the new one, and writes the file back.
func (js *JSONStore) AddDevice(newDevice models.Device) error {
	// First, we read all existing devices
	devices, err := js.GetAllDevices()
	// If the file does not exist or is empty, this is not an error, we just create a new list
	if err != nil && !os.IsNotExist(err) && err.Error() != "EOF" {
		// Check for EOF, as an empty file will return EOF from json.Unmarshal
		if _, ok := err.(*json.SyntaxError); !ok && err.Error() != "EOF" {
             return err
        }
	}


	// Check if a device with this host already exists
	for _, dev := range devices {
		if dev.Host == newDevice.Host {
			return fmt.Errorf("device with host '%s' already exists", newDevice.Host)
		}
	}

	// Add a new device
	devices = append(devices, newDevice)

	// Encode the updated list into JSON with nice formatting
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling devices to JSON: %w", err)
	}

	// Overwrite the file
	err = os.WriteFile(js.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to devices file %s: %w", js.filePath, err)
	}

	return nil
}