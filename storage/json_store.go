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