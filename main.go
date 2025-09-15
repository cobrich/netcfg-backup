package main

import (
	"log"

	"github.com/cobrich/netcfg-backup/cmd"
	"github.com/cobrich/netcfg-backup/monitoring"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	monitoring.StartMetricsServer()

	cmd.Execute()
}
