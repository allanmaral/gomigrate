package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Url            string `yaml:"url"`
	MigrationsPath string `yaml:"migrations-path"`
}

func Init(conf *Config, force bool) error {
	fileName := ".gomigrate"

	if fileExists(fileName) && !force {
		return fmt.Errorf("the file \".gomigrate\" already exists")
	}

	if err := saveConfig(fileName, conf); err != nil {
		return err
	}

	if err := createDirectory(conf.MigrationsPath); err != nil {
		return err
	}

	if err := createGitKeep(conf.MigrationsPath); err != nil {
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
		return fmt.Errorf("failed to create directory \"%s\"", path)
	}

	return nil
}

func createGitKeep(path string) error {
	filename := filepath.Join(path, ".gitkeep")
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create gitignore")
	}
	return nil
}

func saveConfig(filename string, config *Config) error {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to parse config file")
	}

	err = os.WriteFile(filename, configYaml, 0644)
	if err != nil {
		return fmt.Errorf("failed write config file")
	}

	return nil
}
