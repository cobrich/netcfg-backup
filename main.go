/*
netcfg-backup is a tool for concurrently backing up configurations from network
devices using SSH and Telnet. It reads a list of devices from a JSON config,
connects to them in parallel using a worker pool, executes commands, and saves
the output to timestamped files.
*/
package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cobrich/netcfg-backup/config"
	"github.com/cobrich/netcfg-backup/connectors"
	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"

	"github.com/joho/godotenv"
)

const defaultTimeout = 10 * time.Second
const numWorkers = 10

// main is the primary function that orchestrates the backup process.
// It initializes the logger, reads configuration, and starts the worker pool.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	backupPath := flag.String("backup-path", "backups", "Path to the backup directory")
	flag.Parse()

	utils.InitLogger()

	utils.Log.Info("Starting github.com/cobrich/netcfg-backup")

	err = utils.CreateBackup(*backupPath)
	if err != nil {
		utils.Log.WithField("component", "backup").Error(err)
		return
	}

	devices, err := config.ReadConfig()
	if err != nil {
		utils.Log.Fatalf("Failed to load configuration: %v", err)
		
	}

	utils.Log.Infof("Loaded %d devices from configuration", len(devices))
	if len(devices) == 0 {
		utils.Log.Warn("Device list is empty. Exiting.")
		return
	}

	jobs := make(chan models.Device, len(devices))

	var wg sync.WaitGroup

	// A pool of workers is started to process devices concurrently. This prevents overwhelming the network or the host machine.
	utils.Log.Infof("Starting %d workers", numWorkers)
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(&wg, w, jobs, *backupPath)
	}
	utils.Log.Info("All workers started")

	for _, dev := range devices {
		jobs <- dev
	}
	close(jobs)

	wg.Wait()

	utils.Log.Info("All tasks completed.")
}

// worker is a concurrent worker that processes backup jobs from a channel.
// It receives device information, establishes a connection, runs commands, and saves the results.
func worker(wg *sync.WaitGroup, id int, jobs <-chan models.Device, backupPath string) {
	defer wg.Done()

	for dev := range jobs {
		entry := utils.Log.WithFields(map[string]interface{}{
			"worker_id": id,
			"host":      dev.Host,
			"protocol":  dev.Protocol,
		})
		entry.Info("Worker picked up the task")

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
				Host:     dev.Host,
				Username: dev.Username,
				Password: dev.Password,
				KeyPath:  dev.KeyPath,
				Timeout:  timeout,
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
			entry.Error("Unknown protocol")
			continue
		}

		results, err := connector.RunCommands(dev.Commands)
		if err != nil {
			entry.WithField("error", err).Error("Error executing commands")
			continue
		}

		err = utils.WriteResultsToFile(backupPath, dev, results)
		if err != nil {
			entry.WithField("error", err).Error("Error saving results")
		} else {
			entry.Info("Results saved successfully")
		}
	}
}
