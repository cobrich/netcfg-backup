package connectors

import "ssh-fetcher/models"

type Connector interface {
	RunCommands([]string) ([]models.Result, error)
}
