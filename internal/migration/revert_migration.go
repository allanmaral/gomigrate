package migration

import (
	"fmt"
	"time"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/allanmaral/gomigrate/internal/database"
)

func RevertMigration(undoAll bool, conf *config.Config) error {
	driver, err := openDbConnection(conf)
	if err != nil {
		return err
	}
	defer driver.Close()

	appliedMigrations, err := driver.AppliedMigrations()
	if err != nil {
		return err
	}

	if len(appliedMigrations) == 0 {
		fmt.Println("No executed migrations found.")
		return nil
	}

	migrationsCount := len(appliedMigrations)
	for k := range appliedMigrations {
		i := migrationsCount - 1 - k
		migration := appliedMigrations[i]

		err := revertMigration(driver, migration, conf)
		if err != nil {
			return err
		}

		if !undoAll {
			break
		}
	}

	return nil
}

func revertMigration(driver database.Driver, migration string, conf *config.Config) error {
	start := time.Now()
	fmt.Printf("== %s: reverting =======\n", migration)

	mig, err := readMigrationFile(migration, conf)
	if err != nil {
		return err
	}

	if err := driver.Run(mig.Down); err != nil {
		return err
	}

	driver.RemoveApplied(migration)

	elapsed := time.Since(start)
	fmt.Printf("== %s: reverted (%s)\n", migration, elapsed)

	return nil
}
