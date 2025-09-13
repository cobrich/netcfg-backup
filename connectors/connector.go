// Package connectors defines the interface for connecting to network devices and executing commands.
package connectors

import "github.com/cobrich/netcfg-backup/models"

// Connector is the interface that defines the contract for different connection methods (e.g., SSH, Telnet).
type Connector interface {
	// RunCommands executes a list of commands on the device and returns their output.
	RunCommands([]string) ([]models.Result, error)
}
