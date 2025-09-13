package utils

import (
	"fmt"
	"os"
	"path/filepath" // Very important for securely joining paths
	"github.com/cobrich/netcfg-backup/models"
	"time"
)

// CreateBackup now takes a base path
func CreateBackup(basePath string) error {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		// Use MkdirAll, as it doesn't return an error if the folder already exists
		err = os.MkdirAll(basePath, 0755)
		if err != nil {
			Log.Fatalf("❌ Error creating folder %s: %v", basePath, err)
			return err
		}
	}
	return nil
}

// GetDeviceDir now takes a base path
func GetDeviceDir(basePath, host string) (string, error) {
	// Securely join paths: basePath + "/" + host
	dir := filepath.Join(basePath, host)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			Log.Errorf("failed to create folder %s: %v", dir, err)
			return "", err
		}
	}
	return dir, nil
}

// GetBackupFilename now takes a base path
func GetBackupFilename(basePath, host string) (string, error) {
	// First, get the directory for the device, passing basePath
	dir, err := GetDeviceDir(basePath, host)
	if err != nil {
		Log.Error(err)
		return "", err
	}
	tstamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup_%s.txt", tstamp)
	
	// Securely join the directory path and filename
	return filepath.Join(dir, filename), nil
}

// WriteResultsToFile now takes a base path
func WriteResultsToFile(basePath string, dev models.Device, results []models.Result) error {
	// Get the full filename, passing basePath
	filename, err := GetBackupFilename(basePath, dev.Host)
	entry := Log.WithFields(map[string]interface{}{
		"host": dev.Host,
	})
	if err != nil {
		entry.Error(err)
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		entry.WithField("file", filename).Errorf("Failed to create file: %v", err)
		return err
	}
	defer f.Close()

	// file header
	f.WriteString("########################################\n")
	f.WriteString(fmt.Sprintf(" Host: %s\n", dev.Host))
	f.WriteString(fmt.Sprintf(" User: %s\n", dev.Username))
	f.WriteString(fmt.Sprintf(" Date: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString("########################################\n\n")

	// command output
	for _, result := range results {
		f.WriteString("### " + result.Cmd + " ###\n")
		f.WriteString(result.Output + "\n\n")
	}

	entry.WithField("file", filename).Info("✅ Result saved")
	return nil
}
