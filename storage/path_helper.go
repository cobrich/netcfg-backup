package storage

import (
	"os"
	"path/filepath"
)

// GetDefaultDBPath returns the default path for the SQLite database file.
// It ensures the parent directory exists.
// Path: ~/.config/netcfg-backup/netcfg.db
func GetDefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(home, ".config", "netcfg-backup")
	
	// Create the directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}
	
	return filepath.Join(configDir, "netcfg.db"), nil
}