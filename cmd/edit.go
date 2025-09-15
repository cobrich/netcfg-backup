package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [host]",
	Short: "Interactively edit an existing device in the configuration",
	Long:  `Allows you to interactively edit the details of a device specified by its host.
For each field, the current value is displayed. Press Enter to keep the current value, or type a new one to change it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostToEdit := args[0]
		deviceStore := storage.NewJSONStore("devices/devices.json")

		// 1. Find the device to edit
		device, err := deviceStore.GetDeviceByHost(hostToEdit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("--- Editing device '%s' ---\n", device.Host)
		fmt.Println("(Press Enter to keep the current value)")

		// 2. Interactively ask for new values, showing the old ones as defaults
		device.Username = askQuestionWithDefault(reader, "Username", device.Username)

		// Edit Protocol
		// Note: Changing protocol might invalidate auth method, so we handle that.
		newProtocol := askChoiceWithDefault(reader, "Protocol (ssh/telnet)", []string{"ssh", "telnet"}, device.Protocol)
		protocolChanged := newProtocol != device.Protocol
		device.Protocol = newProtocol

		// Edit Auth Method (conditionally)
		if device.Protocol == "ssh" {
			currentAuthMethod := "password"
			if device.KeyPath != "" {
				currentAuthMethod = "key"
			}
			newAuthMethod := askChoiceWithDefault(reader, "Authentication method (key/password)", []string{"key", "password"}, currentAuthMethod)

			if newAuthMethod == "key" {
				defaultKeyPath := device.KeyPath
				if defaultKeyPath == "" {
					defaultKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
				}
				device.KeyPath = askQuestionWithDefault(reader, "Path to SSH key file", defaultKeyPath)
				device.Password = "" // Clear password fields if key is used
				device.PasswordEnv = ""
			} else { // password
				device.PasswordEnv = askQuestionWithDefault(reader, "Environment variable for the password", device.PasswordEnv)
				device.KeyPath = "" // Clear key path if password is used
			}
		} else { // telnet
			device.PasswordEnv = askQuestionWithDefault(reader, "Environment variable for the password", device.PasswordEnv)
			device.Prompt = askQuestionWithDefault(reader, "Telnet prompt symbol", device.Prompt)
			device.KeyPath = "" // Clear key path for Telnet
		}
		
		// If protocol changed from ssh to telnet, we might need a prompt
		if protocolChanged && device.Protocol == "telnet" && device.Prompt == "" {
			device.Prompt = askQuestionWithDefault(reader, "Telnet prompt symbol", "#")
		}

		// Edit Commands
		fmt.Printf("Current commands: %v\n", device.Commands)
		if askChoice(reader, "Do you want to re-enter all commands?", []string{"yes", "no"}) == "yes" {
			device.Commands = []string{} // Clear existing commands
			fmt.Println("Enter new commands, one per line. Type 'done' when finished.")
			for {
				newCmd := askQuestion(reader, "> ")
				if strings.ToLower(newCmd) == "done" {
					break
				}
				device.Commands = append(device.Commands, newCmd)
			}
		}

		// 3. Save the updated device
		if err := deviceStore.UpdateDevice(*device); err != nil {
			fmt.Printf("Error updating device: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Device '%s' updated successfully!\n", device.Host)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

// Helper function to ask for a choice with a default value
func askChoiceWithDefault(reader *bufio.Reader, query string, choices []string, defaultValue string) string {
	for {
		// Highlight the default choice
		promptChoices := make([]string, len(choices))
		for i, c := range choices {
			if c == defaultValue {
				promptChoices[i] = strings.ToUpper(c)
			} else {
				promptChoices[i] = c
			}
		}

		fmt.Printf("%s (%s) [%s]: ", query, strings.Join(promptChoices, "/"), defaultValue)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultValue
		}

		for _, choice := range choices {
			if input == choice {
				return input
			}
		}
		fmt.Printf("Invalid choice. Please select one of: %s\n", strings.Join(choices, ", "))
	}
}