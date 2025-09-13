package connectors

import "github.com/cobrich/netcfg-backup/models"

type Connector interface {
	RunCommands([]string) ([]models.Result, error)
}
