package cmd

import (
	"fmt"
	"os"

	"github.com/cobrich/netcfg-backup/backups"
	"github.com/cobrich/netcfg-backup/core"
	"github.com/cobrich/netcfg-backup/server"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the web interface",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, err := storage.GetDefaultDBPath()
		if err != nil {
			fmt.Printf("Error determining database path: %v\n", err)
			os.Exit(1)
		}
		deviceStore, err := storage.NewSQLiteStore(dbPath)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			os.Exit(1)
		}

		backupPath := "backups"
		backupSvc := backups.NewService(backupPath)
		coreSvc := core.NewBackupService(deviceStore, backupPath, 10)

		srv := server.New(deviceStore, backupSvc, coreSvc)
		srv.Start("localhost:8080")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
