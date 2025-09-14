// cmd/list.go
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
		// Создаем экземпляр нашего хранилища
		deviceStore := storage.NewJSONStore("devices/devices.json")

		// Получаем все устройства
		devices, err := deviceStore.GetAllDevices()
		if err != nil {
			fmt.Printf("Error loading devices: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("No devices configured. Use 'netcfg-backup add' to add one.")
			return
		}

		// Выводим красивую шапку таблицы
		fmt.Printf("%-20s %-15s %-10s %s\n", "HOST", "USERNAME", "PROTOCOL", "AUTH METHOD")
		fmt.Println("-----------------------------------------------------------------")

		// Проходим по устройствам и печатаем информацию
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