package cmd

import (
	"github.com/allanmaral/gomigrate/internal/migration"
	"github.com/spf13/cobra"
)

// migrationDownCmd represents the revert migration command
var migrationDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Revert last applied migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		if err := migration.RevertMigration(config); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrationDownCmd)
}
