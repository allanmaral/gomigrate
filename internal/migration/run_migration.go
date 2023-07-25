package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/allanmaral/gomigrate/internal/database"
	_ "github.com/go-sql-driver/mysql"
)

type Migration struct {
	Up   string
	Down string
}

func RunMigrations(conf *config.Config) error {
	driver, err := openDbConnection(conf)
	if err != nil {
		return err
	}
	defer driver.Close()

	appliedMigrations, err := driver.AppliedMigrations()
	if err != nil {
		return err
	}

	localMigrations, err := loadMigrationScripts(conf.MigrationsPath)
	if err != nil {
		return err
	}

	missingMigrations := findMissingMigrations(appliedMigrations, localMigrations)

	if len(missingMigrations) == 0 {
		fmt.Println("No migrations were executed, database schema was already up to date.")
		return nil
	}

	for _, migration := range missingMigrations {
		fmt.Printf("== %s: migrating =======\n", migration)

		mig, err := readMigrationFile(migration, conf)
		if err != nil {
			return err
		}

		if err := driver.Run(mig.Up); err != nil {
			return err
		}

		driver.MarkAsApplied(migration)

		fmt.Printf("== %s: migrated (?s)\n", migration)
	}

	return nil
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
		return nil, errors.New("failed to scan migrations folder")
	}

	matchingFiles := []string{}
	for _, file := range files {
		match, err := filepath.Match(pattern, file.Name())
		if err != nil {
			return nil, errors.New("failed to match migration names")
		}

		if match {
			matchingFiles = append(matchingFiles, file.Name())
		}
	}

	return matchingFiles, nil
}

func findMissingMigrations(applied []string, available []string) []string {
	missing := []string{}
	appliedMap := make(map[string]bool, len(applied))
	for _, kApplied := range applied {
		appliedMap[kApplied] = true
	}
	for _, kAvailable := range available {
		if !appliedMap[kAvailable] {
			missing = append(missing, kAvailable)
		}
	}

	sort.Strings(missing)

	return missing
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
