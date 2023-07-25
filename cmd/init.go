package cmd

import (
	"fmt"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/allanmaral/gomigrate/internal/database"
	"github.com/spf13/cobra"
)

var (
	username       string
	password       string
	databaseName   string
	hostname       string
	port           int32
	provider       string
	migrationsPath string
	force          bool

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes project",
		Long:  "Initializes project",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &database.ConnectionParams{
				User:     username,
				Password: password,
				Database: databaseName,
				Hostname: hostname,
				Port:     port,
				Provider: provider,
			}

			url, err := database.Url(params)
			if err != nil {
				return fmt.Errorf("failed to make connection url")
			}

			conf := config.Config{
				Url:            url.String(),
				MigrationsPath: migrationsPath,
			}

			if err := config.Init(&conf, force); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	initCmd.Flags().StringVarP(&password, "password", "p", "password", "Database password")
	initCmd.Flags().StringVarP(&databaseName, "database", "d", "", "Database name")
	initCmd.Flags().StringVar(&hostname, "host", "localhost", "Database connection hostname")
	initCmd.Flags().Int32Var(&port, "port", 0, "Database connection port")
	initCmd.Flags().StringVar(&provider, "provider", "sqlserver", "Database provider")
	initCmd.Flags().StringVar(&migrationsPath, "migrations-path", "migrations", "Path to the migrations folder")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Will drop the existing config file and re-create it")
}
