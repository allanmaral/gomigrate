package cmd

import (
	"fmt"
	"os"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "gomigrate",
		Short: "A CLI for managing database migrations",
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func GetConfig() *config.Config {
	config := &config.Config{
		Username:       viper.GetString("username"),
		Password:       viper.GetString("password"),
		Database:       viper.GetString("database"),
		Host:           viper.GetString("host"),
		Dialect:        viper.GetString("dialect"),
		MigrationsPath: viper.GetString("migrations-path"),
	}

	return config
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .gomigrate)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".gomigrate" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath("../")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gomigrate")
	}

	viper.SetDefault("username", "root")
	viper.SetDefault("password", "")
	viper.SetDefault("database", "database")
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("dialect", "postgres")
	viper.SetDefault("migrations_path", "migrations")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
