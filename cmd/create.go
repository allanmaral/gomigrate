package cmd

import (
	"github.com/allanmaral/gomigrate/internal/migration"
	"github.com/spf13/cobra"
)

var migrationName string

// migrationCreateCmd represents the create migration command
var migrationCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Generates a new migration file",
	Args:  ArgsValidator,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		if err := migration.NewMigration(migrationName, config); err != nil {
			return err
		}

		return nil
	},
}

func ArgsValidator(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	if len(name) == 0 {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		migrationName = args[0]
	} else {
		migrationName = name
	}

	return nil
}

func init() {
	rootCmd.AddCommand(migrationCreateCmd)

	migrationCreateCmd.Flags().StringP("name", "n", "", "Defines the name of the migration (required)")
}
