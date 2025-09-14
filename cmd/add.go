package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactively add a new device to the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)
		newDevice := models.Device{}

		fmt.Println("--- Adding a new device ---")

		// Ask for data
		newDevice.Host = askQuestion(reader, "Enter hostname or IP address: ")
		newDevice.Username = askQuestion(reader, "Enter username: ")
		
		protocol := askChoice(reader, "Select protocol:", []string{"ssh", "telnet"})
		newDevice.Protocol = protocol

		if protocol == "ssh" {
			authMethod := askChoice(reader, "Select authentication method:", []string{"key", "password"})
			if authMethod == "key" {
				defaultKeyPath := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
				newDevice.KeyPath = askQuestionWithDefault(reader, "Enter path to SSH key file:", defaultKeyPath)
			} else {
				newDevice.PasswordEnv = askQuestion(reader, "Enter environment variable name for the password: ")
			}
		} else { // telnet
			newDevice.PasswordEnv = askQuestion(reader, "Enter environment variable name for the password: ")
			newDevice.Prompt = askQuestionWithDefault(reader, "Enter Telnet prompt symbol:", "#")
		}

		// Add the device through our storage
		deviceStore := storage.NewJSONStore("devices/devices.json")
		if err := deviceStore.AddDevice(newDevice); err != nil {
			fmt.Printf("Error adding device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Device '%s' added successfully!\n", newDevice.Host)
		fmt.Println("Don't forget to set the password in your .env file if needed.")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// Helper functions for polling the user

func askQuestion(reader *bufio.Reader, query string) string {
	for {
		fmt.Print(query)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Println("This field cannot be empty.")
	}
}

func askQuestionWithDefault(reader *bufio.Reader, query, defaultValue string) string {
	fmt.Printf("%s [%s]: ", query, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func askChoice(reader *bufio.Reader, query string, choices []string) string {
	for {
		fmt.Printf("%s (%s): ", query, strings.Join(choices, "/"))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		for _, choice := range choices {
			if input == choice {
				return input
			}
		}
		fmt.Printf("Invalid choice. Please select one of: %s\n", strings.Join(choices, ", "))
	}
}
