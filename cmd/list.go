package cmd

import (
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all configured devices",
	Run: func(cmd *cobra.Command, args []string) {
		// Create an instance of our storage
		// deviceStore := storage.NewJSONStore("devices/devices.json")
		// Create a path to the database file in the user's home directory
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

		// Get all devices

		devices, err := deviceStore.GetAllDevices()
		if err != nil {
			fmt.Printf("Error loading devices: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("No devices configured. Use 'netcfg-backup add' to add one.")
			return
		}

		// Print a nice table header
		fmt.Printf("%-20s %-15s %-10s %s\n", "HOST", "USERNAME", "PROTOCOL", "AUTH METHOD")
		fmt.Println("-----------------------------------------------------------------")

		// Loop through the devices and print the information
		for _, dev := range devices {
			authMethod := "Password"
			if dev.KeyPath != "" {
				authMethod = "SSH Key"
			}
			fmt.Printf("%-20s %-15s %-10s %s\n", dev.Host, dev.Username, dev.Protocol, authMethod)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
