package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	Host           string `yaml:"host"`
	Dialect        string `yaml:"dialect"`
	MigrationsPath string `yaml:"migrations-path"`
}

func Init(config *Config, force bool) error {
	fileName := ".gomigrate"

	if fileExists(fileName) && !force {
		return errors.New("the file \".gomigrate\" already exists")
	}

	if err := saveConfig(fileName, config); err != nil {
		return err
	}

	if err := createDirectory(config.MigrationsPath); err != nil {
		return err
	}

	if err := createGitKeep(config.MigrationsPath); err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createDirectory(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		message := fmt.Sprintf("failed to create directory \"%s\"", path)
		return errors.New(message)
	}

	return nil
}

func createGitKeep(path string) error {
	filename := filepath.Join(path, ".gitkeep")
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		return errors.New("failed to create gitignore")
	}
	return nil
}

// TODO: Move to infra layer
func saveConfig(filename string, config *Config) error {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return errors.New("failed to parse config file")
	}

	err = os.WriteFile(filename, configYaml, 0644)
	if err != nil {
		return errors.New("failed write config file")
	}

	return nil
}
