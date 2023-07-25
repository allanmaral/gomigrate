package migration

import (
	"fmt"

	"github.com/allanmaral/gomigrate/internal/config"
)

func RevertMigration(conf *config.Config) error {
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

	lastMigration := appliedMigrations[len(appliedMigrations)-1]

	fmt.Printf("== %s: reverting =======\n", lastMigration)

	mig, err := readMigrationFile(lastMigration, conf)
	if err != nil {
		return err
	}

	if err := driver.Run(mig.Down); err != nil {
		return err
	}

	driver.RemoveApplied(lastMigration)

	fmt.Printf("== %s: reverted (?s)\n", lastMigration)

	return nil
}
