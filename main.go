package main

import (
	"github.com/cobrich/netcfg-backup/cmd"
	"github.com/cobrich/netcfg-backup/monitoring"
)

func main() {
	monitoring.StartMetricsServer()

	cmd.Execute()
}
