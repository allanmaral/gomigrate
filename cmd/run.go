package cmd

import (
	"github.com/allanmaral/gomigrate/internal/migration"
	"github.com/spf13/cobra"
)

// migrationCreateCmd represents the create migration command
var migrationRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		if err := migration.RunMigrations(config); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrationRunCmd)
}
