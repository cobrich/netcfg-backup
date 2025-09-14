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
	// Сначала читаем все существующие устройства
	devices, err := js.GetAllDevices()
	// Если файл не существует или пуст, это не ошибка, просто создаем новый список
	if err != nil && !os.IsNotExist(err) && err.Error() != "EOF" {
		// Проверяем на EOF, так как пустой файл json.Unmarshal вернет EOF
		if _, ok := err.(*json.SyntaxError); !ok && err.Error() != "EOF" {
             return err
        }
	}


	// Проверяем, не существует ли уже устройство с таким хостом
	for _, dev := range devices {
		if dev.Host == newDevice.Host {
			return fmt.Errorf("device with host '%s' already exists", newDevice.Host)
		}
	}

	// Добавляем новое устройство
	devices = append(devices, newDevice)

	// Кодируем обновленный список в JSON с красивым форматированием
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling devices to JSON: %w", err)
	}

	// Перезаписываем файл
	err = os.WriteFile(js.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to devices file %s: %w", js.filePath, err)
	}

	return nil
}