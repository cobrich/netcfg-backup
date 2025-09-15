package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		// If error - "file not found"
		if os.IsNotExist(err) {
			// Check, if exists folder
			dir := filepath.Dir(js.filePath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}
			
			// Create empty file
			if err := os.WriteFile(js.filePath, []byte("[]\n"), 0644); err != nil {
				return nil, fmt.Errorf("failed to create empty devices file %s: %w", js.filePath, err)
			}
			// Return empty list of devices, not error
			return []models.Device{}, nil
		}
		// If another error, return it
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

// RemoveDevice removes a device from the JSON file by its host.
func (js *JSONStore) RemoveDevice(host string) error {
	devices, err := js.GetAllDevices()
	if err != nil {
		return err
	}

	// Create a new slice to store the devices we want to keep
	var updatedDevices []models.Device
	found := false

	for _, dev := range devices {
		if dev.Host == host {
			found = true // We have found a device that needs to be removed.
		} else {
			// We are keeping this device and adding it to a new list.
			updatedDevices = append(updatedDevices, dev)
		}
	}

	// If we still haven't found the device, we will notify the user.
	if !found {
		return fmt.Errorf("device with host '%s' not found", host)
	}

	// Encode the new (reduced) list in JSON
	data, err := json.MarshalIndent(updatedDevices, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling devices to JSON: %w", err)
	}

	// Rewrite the file
	err = os.WriteFile(js.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to devices file %s: %w", js.filePath, err)
	}

	return nil
}

// GetDeviceByHost finds a single device by its host.
func (js *JSONStore) GetDeviceByHost(host string) (*models.Device, error) {
	devices, err := js.GetAllDevices()
	if err != nil {
		return nil, err
	}
	for i, dev := range devices {
		if dev.Host == host {
			return &devices[i], nil // Возвращаем указатель на элемент в слайсе
		}
	}
	return nil, fmt.Errorf("device with host '%s' not found", host)
}

// UpdateDevice finds a device by its host and replaces it with the new version.
func (js *JSONStore) UpdateDevice(updatedDevice models.Device) error {
	devices, err := js.GetAllDevices()
	if err != nil {
		return err
	}

	found := false
	for i, dev := range devices {
		if dev.Host == updatedDevice.Host {
			devices[i] = updatedDevice // Заменяем старую структуру на новую
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("device with host '%s' not found to update", updatedDevice.Host)
	}

	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling devices: %w", err)
	}

	return os.WriteFile(js.filePath, data, 0644)
}