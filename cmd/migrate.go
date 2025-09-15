package cmd

import (
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrates devices from devices.json to the SQLite database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting migration from devices.json to SQLite...")

		jsonStore := storage.NewJSONStore("devices/devices.json")
		devices, err := jsonStore.GetAllDevices()
		if err != nil {
			fmt.Printf("Could not read from devices.json: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("devices.json is empty. Nothing to migrate.")
			return
		}

		dbPath, _ := storage.GetDefaultDBPath()
		sqliteStore, err := storage.NewSQLiteStore(dbPath)
		if err != nil {
			fmt.Printf("Could not open SQLite database: %v\n", err)
			os.Exit(1)
		}

		count := 0
		for _, dev := range devices {
			if err := sqliteStore.AddDevice(dev); err != nil {
				if _, ok := err.(*storage.ErrDeviceExists); ok {
					fmt.Printf("Device %s already exists in database, skipping.\n", dev.Host)
				} else {
					fmt.Printf("Failed to add device %s: %v\n", dev.Host, err)
				}
			} else {
				fmt.Printf("Migrated device: %s\n", dev.Host)
				count++
			}
		}

		fmt.Printf("\n✅ Migration complete. %d new devices were added to the database.\n", count)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}