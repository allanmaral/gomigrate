package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/allanmaral/gomigrate/internal/database/sqlserver"
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
	configDir := filepath.Dir(viper.ConfigFileUsed())
	migrationsPath := path.Join(configDir, viper.GetString("migrations_path"))

	config := &config.Config{
		Url:            viper.GetString("url"),
		MigrationsPath: migrationsPath,
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

	viper.SetDefault("url", "postgres://postgres:password@localhost:5432/example?sslmode=disable")
	viper.SetDefault("migrations-path", "migrations")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path, err := filepath.Rel(wd, viper.ConfigFileUsed())

		if err != nil {
			fmt.Printf("Loaded configuration file \"%s\"\n", viper.ConfigFileUsed())
		} else {
			fmt.Printf("Loaded configuration file \"%s\"\n", path)
		}
	}
}
