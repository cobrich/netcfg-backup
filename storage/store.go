// Package storage defines the interface for data persistence.
package storage

import "github.com/cobrich/netcfg-backup/models"

// Store is the interface for device storage.
// It abstracts the underlying storage mechanism (e.g., JSON file, database).
type Store interface {
	GetAllDevices() ([]models.Device, error)
	AddDevice(device models.Device) error
	RemoveDevice(host string) error
	GetDeviceByHost(host string) (*models.Device, error)
	UpdateDevice(updatedDevice models.Device) error
}
