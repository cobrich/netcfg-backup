package cmd

import (
	// ... импорты, которые раньше были в main.go ...
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cobrich/netcfg-backup/connectors"
	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/joho/godotenv"

	"github.com/spf13/cobra"
)

const defaultTimeout = 10 * time.Second
const numWorkers = 10

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the backup process for all configured devices",
	Run: func(cmd *cobra.Command, args []string) {
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

		// 1. Создаем экземпляр нашего JSON-хранилища.
		deviceStore := storage.NewJSONStore("devices/devices.json")

		// 2. Получаем устройства, используя наше новое хранилище.
		devices, err := deviceStore.GetAllDevices()
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
	},
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

func init() {
	rootCmd.AddCommand(runCmd)

	// Здесь мы можем определять флаги, специфичные для команды 'run'
	// Например, тот самый --backup-path
	runCmd.Flags().StringP("backup-path", "p", "backups", "Path to the backup directory")
}
