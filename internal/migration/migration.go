package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/allanmaral/gomigrate/internal/database"
)

type Migration struct {
	Up   string
	Down string
}

func openDbConnection(conf *config.Config) (database.Driver, error) {
	driver, err := database.Open(conf.Url)
	if err != nil {
		return nil, err
	}

	return driver, nil
}

func loadMigrationScripts(path string) ([]string, error) {
	pattern := "*.sql"
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan migrations folder")
	}

	matchingFiles := []string{}
	for _, file := range files {
		match, err := filepath.Match(pattern, file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to match migration names")
		}

		if match {
			matchingFiles = append(matchingFiles, file.Name())
		}
	}

	return matchingFiles, nil
}

func readMigrationFile(migration string, conf *config.Config) (*Migration, error) {
	dat, err := os.ReadFile(filepath.Join(conf.MigrationsPath, migration))
	if err != nil {
		return nil, err
	}

	fileStr := string(dat)
	sections := strings.Split(fileStr, "BEGIN -- UP")
	if len(sections) < 2 {
		return nil, fmt.Errorf("could not find start of UP section in file %s", migration)
	}

	fileStr = sections[1]
	sections = strings.Split(fileStr, "END -- UP")
	if len(sections) < 2 {
		return nil, fmt.Errorf("could not find end of UP section in file %s", migration)
	}

	up := sections[0]

	fileStr = sections[1]
	sections = strings.Split(fileStr, "BEGIN -- DOWN")
	if len(sections) < 2 {
		return nil, fmt.Errorf("could not find start of DOWN section in file %s", migration)
	}

	fileStr = sections[1]
	sections = strings.Split(fileStr, "END -- DOWN")
	if len(sections) < 2 {
		return nil, fmt.Errorf("could not find end of DOWN section in file %s", migration)
	}

	down := sections[0]

	mig := &Migration{
		Up:   up,
		Down: down,
	}

	return mig, nil
}
