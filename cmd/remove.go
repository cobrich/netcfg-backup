package cmd

import (
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [host]",
	Short: "Removes a device from the configuration",
	Long:  `Removes a device specified by its host name or IP address from the devices.json file.`,
	// Args guarantees that the command will be called with exactly one argument.
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostToRemove := args[0]

		deviceStore := storage.NewJSONStore("devices/devices.json")

		if err := deviceStore.RemoveDevice(hostToRemove); err != nil {
			fmt.Printf("Error removing device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Device '%s' removed successfully!\n", hostToRemove)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}