package cmd

import (
	"github.com/allanmaral/gomigrate/internal/migration"
	"github.com/spf13/cobra"
)

var revertAll bool

// migrationUndoCmd represents the revert migration command
var migrationUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Revert applied migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		if err := migration.RevertMigration(revertAll, config); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrationUndoCmd)

	migrationUndoCmd.Flags().BoolVarP(&revertAll, "all", "a", false, "Revert all migrations")
}
