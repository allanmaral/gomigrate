package migration

import (
	"fmt"
	"sort"
	"time"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/allanmaral/gomigrate/internal/database"
)

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
		if err := runMigration(driver, migration, conf); err != nil {
			return err
		}
	}

	return nil
}

func runMigration(driver database.Driver, migration string, conf *config.Config) error {
	start := time.Now()
	fmt.Printf("== %s: migrating =======\n", migration)

	mig, err := readMigrationFile(migration, conf)
	if err != nil {
		return err
	}

	if err := driver.Run(mig.Up); err != nil {
		return err
	}

	if err := driver.MarkAsApplied(migration); err != nil {
		return err
	}

	elapsed := time.Since(start)
	fmt.Printf("== %s: migrated (%s)\n", migration, elapsed)

	return nil
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
