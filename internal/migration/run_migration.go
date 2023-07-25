package migration

import (
	"fmt"
	"sort"

	"github.com/allanmaral/gomigrate/internal/config"
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
