package cmd

import (
	"fmt"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var force bool
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes project",
	Long:  "Initializes project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("init called")

		conf := GetConfig()
		if err := config.Init(conf, force); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("username", "u", "", "Database username")
	initCmd.Flags().StringP("password", "p", "", "Database password")
	initCmd.Flags().StringP("database", "d", "", "Database name")
	initCmd.Flags().String("host", "", "Database connection url")
	initCmd.Flags().String("dialect", "", "Database dialect")
	initCmd.Flags().String("migrations-path", "", "Path to the migrations folder")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Will drop the existing config file and re-create it")

	viper.BindPFlag("username", initCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", initCmd.Flags().Lookup("password"))
	viper.BindPFlag("database", initCmd.Flags().Lookup("database"))
	viper.BindPFlag("host", initCmd.Flags().Lookup("host"))
	viper.BindPFlag("dialect", initCmd.Flags().Lookup("dialect"))
	viper.BindPFlag("migrations-path", initCmd.Flags().Lookup("migrations-path"))
}
