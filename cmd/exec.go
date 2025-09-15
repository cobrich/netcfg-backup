package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cobrich/netcfg-backup/connectors"
	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute commands on a single device without saving it to the inventory",
	Long: `The exec command allows for ad-hoc command execution on a single network device.
All connection parameters must be provided via command-line flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.InitLogger()

		// Get alll data from flags
		host, _ := cmd.Flags().GetString("host")
		username, _ := cmd.Flags().GetString("username")
		protocol, _ := cmd.Flags().GetString("protocol")
		keyPath, _ := cmd.Flags().GetString("key-path")
		passwordEnv, _ := cmd.Flags().GetString("password-env")
		commands, _ := cmd.Flags().GetStringSlice("command")
		timeoutSeconds, _ := cmd.Flags().GetInt("timeout")
		timeout := time.Duration(timeoutSeconds) * time.Second
		allowInsecure, _ := cmd.Flags().GetBool("insecure-algos")
		telnetPrompt, _ := cmd.Flags().GetString("prompt")

		// Simple validation
		if host == "" || username == "" || len(commands) == 0 {
			fmt.Println("Error: --host, --username, and at least one --command are required.")
			os.Exit(1)
		}

		// Create only one Device object
		device := models.Device{
			Host:     host,
			Username: username,
			Protocol: protocol,
			KeyPath:  keyPath,
			Commands: commands,
		}

		if passwordEnv != "" {
			device.Password = os.Getenv(passwordEnv)
		}

		// Run lpgic for connecting (simple version of worker)
		entry := utils.Log.WithFields(map[string]interface{}{
			"host":     device.Host,
			"protocol": device.Protocol,
		})

		var connector connectors.Connector

		switch device.Protocol {
		case "ssh":
			connector = &connectors.SSHConnector{
				Host:               device.Host,
				Username:           device.Username,
				Password:           device.Password,
				KeyPath:            device.KeyPath,
				Timeout:            timeout,
				AllowInsecureAlgos: allowInsecure,
			}
		case "telnet":
			connector = &connectors.TelnetConnector{
				Host:     device.Host,
				Username: device.Username,
				Password: device.Password,
				Prompt:   telnetPrompt,
				Timeout:  timeout,
			}
		default:
			entry.Errorf("Unknown protocol: %s", device.Protocol)
			os.Exit(1)
		}

		results, err := connector.RunCommands(device.Commands)
		if err != nil {
			entry.Errorf("Error executing commands: %v", err)
			os.Exit(1)
		}

		fmt.Printf("--- Results for %s ---\n", device.Host)
		for _, result := range results {
			fmt.Printf("\n### Command: %s ###\n", result.Cmd)
			fmt.Println(result.Output)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Define flags for exec
	execCmd.Flags().String("host", "", "Target device hostname or IP address (required)")
	execCmd.Flags().String("username", "", "Username for authentication (required)")
	execCmd.Flags().String("protocol", "ssh", "Connection protocol (ssh or telnet)")
	execCmd.Flags().String("key-path", "", "Path to SSH private key file")
	execCmd.Flags().String("password-env", "", "Environment variable for the password")
	execCmd.Flags().StringSlice("command", []string{}, "Command to execute (required, can be specified multiple times)")
	execCmd.Flags().Int("timeout", 15, "Connection timeout in seconds")
	execCmd.Flags().Bool("insecure-algos", false, "Allow insecure legacy SSH algorithms")
	execCmd.Flags().String("prompt", "#", "Telnet prompt symbol to expect")
}
