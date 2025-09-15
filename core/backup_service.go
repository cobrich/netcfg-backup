// Package core contains the main business logic of the application.
package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cobrich/netcfg-backup/connectors"
	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/monitoring"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/cobrich/netcfg-backup/utils"
)

const defaultTimeout = 10 * time.Second

// BackupService orchestrates the backup process.
type BackupService struct {
	store      storage.Store
	basePath   string
	numWorkers int
}

// NewBackupService creates a new backup service.
func NewBackupService(store storage.Store, basePath string, numWorkers int) *BackupService {
	return &BackupService{
		store:      store,
		basePath:   basePath,
		numWorkers: numWorkers,
	}
}

// Run executes the backup process for all devices.
// It runs in the foreground and returns when all jobs are complete.
func (s *BackupService) Run() error {
	utils.Log.Info("Starting backup run...")

	devices, err := s.store.GetAllDevices()
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	if len(devices) == 0 {
		utils.Log.Warn("Device list is empty. Nothing to do.")
		return nil
	}
	utils.Log.Infof("Loaded %d devices from configuration", len(devices))

	jobs := make(chan models.Device, len(devices))
	var wg sync.WaitGroup

	utils.Log.Infof("Starting %d workers", s.numWorkers)
	for w := 1; w <= s.numWorkers; w++ {
		wg.Add(1)
		// Мы передаем s.basePath в воркера
		go s.worker(&wg, w, jobs)
	}

	for _, dev := range devices {
		jobs <- dev
	}
	close(jobs)

	wg.Wait()
	utils.Log.Info("All backup tasks completed.")
	return nil
}

// worker function is now a method of BackupService
func (s *BackupService) worker(wg *sync.WaitGroup, id int, jobs <-chan models.Device) {
	defer wg.Done()

	for dev := range jobs {
		// Set starting time
		startTime := time.Now()

		entry := utils.Log.WithFields(map[string]interface{}{
			"worker_id": id,
			"host":      dev.Host,
			"protocol":  dev.Protocol,
		})
		entry.Info("Worker picked up the task")

		status := "success"
		var finalErr error

		func() {
			if dev.PasswordEnv != "" {
				dev.Password = os.Getenv(dev.PasswordEnv)
				if dev.Password == "" {
					entry.Warnf("Environment variable '%s' is not set or empty", dev.PasswordEnv)
				}
			}

			timeout := defaultTimeout
			if dev.TimeoutSeconds > 0 {
				timeout = time.Duration(dev.TimeoutSeconds) * time.Second
			}

			var connector connectors.Connector
			switch dev.Protocol {
			case "ssh":
				connector = &connectors.SSHConnector{
					Host:               dev.Host,
					Username:           dev.Username,
					Password:           dev.Password,
					KeyPath:            dev.KeyPath,
					Timeout:            timeout,
					AllowInsecureAlgos: dev.AllowInsecureAlgos,
				}
			case "telnet":
				connector = &connectors.TelnetConnector{
					Host:     dev.Host,
					Username: dev.Username,
					Password: dev.Password,
					Prompt:   dev.Prompt,
					Timeout:  timeout,
				}
			default:
				finalErr = fmt.Errorf("unknown protocol: %s", dev.Protocol)
				entry.Error("Unknown protocol")
				return
			}

			results, err := connector.RunCommands(dev.Commands)
			if err != nil {
				finalErr = err
				entry.WithField("error", finalErr).Error("Error executing commands")
				return
			}

			if err := utils.WriteResultsToFile(s.basePath, dev, results); err != nil {
				finalErr = err
				entry.WithField("error", finalErr).Error("Error saving results")
			} else {
				entry.Info("Results saved successfully")
			}
		}()

		duration := time.Since(startTime).Seconds()
		if finalErr != nil {
			status = "failed"
		}

		monitoring.JobsTotal.WithLabelValues(dev.Host, status).Inc()
		monitoring.JobDuration.WithLabelValues(dev.Host).Observe(duration)

		entry.Infof("Job finished with status '%s' in %.2f seconds", status, duration)
	}
}