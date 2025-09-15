package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/core"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/cobrich/netcfg-backup/utils"

	"github.com/spf13/cobra"
)

const numWorkers = 10

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the backup process for all configured devices",
	Run: func(cmd *cobra.Command, args []string) {

		backupPath := flag.String("backup-path", "backups", "Path to the backup directory")
		flag.Parse()

		utils.InitLogger()

		utils.Log.Info("Starting github.com/cobrich/netcfg-backup")

		err := utils.CreateBackup(*backupPath)
		if err != nil {
			utils.Log.WithField("component", "backup").Error(err)
			return
		}

		// deviceStore := storage.NewJSONStore("devices/devices.json")
		dbPath, err := storage.GetDefaultDBPath()
		if err != nil {
			fmt.Printf("Error determining database path: %v\n", err)
			os.Exit(1)
		}
		deviceStore, err := storage.NewSQLiteStore(dbPath)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			os.Exit(1)
		}

		devices, err := deviceStore.GetAllDevices()
		if err != nil {
			utils.Log.Fatalf("Failed to load configuration: %v", err)
		}

		utils.Log.Infof("Loaded %d devices from configuration", len(devices))
		if len(devices) == 0 {
			utils.Log.Warn("Device list is empty. Exiting.")
			return
		}

		backupService := core.NewBackupService(deviceStore, *backupPath, 10) // 10 - numWorkers
		if err := backupService.Run(); err != nil {
			utils.Log.Fatalf("Backup process failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here we can define flags specific to the 'run' command
	// For example, the same --backup-path
	runCmd.Flags().StringP("backup-path", "p", "backups", "Path to the backup directory")
}
