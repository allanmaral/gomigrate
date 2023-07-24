package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/allanmaral/gomigrate/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

func RunMigrations(c *config.Config) error {
	db, err := openDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := createMetaTable(db); err != nil {
		return err
	}

	appliedMigrations, err := loadAppliedMigrations(db)
	if err != nil {
		return err
	}

	localMigrations, err := loadMigrationScripts(c.MigrationsPath)
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

		// apply migration

		markAsApplied(migration, db)

		fmt.Printf("== %s: migrated (?s)\n", migration)
	}

	return nil
}

func openDbConnection() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:my-secret-pw@tcp(127.0.0.1:3306)/database")
	if err != nil {
		return nil, errors.New("failed to connect to the database")
	}

	return db, nil
}

func createMetaTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gomigrate_meta 
		(
			name VARCHAR(255) NOT NULL
				PRIMARY KEY,
			CONSTRAINT name 
				UNIQUE (name)
		);
	`)
	if err != nil {
		return errors.New("failed to create metadata table")
	}

	return nil
}

func loadAppliedMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM gomigrate_meta;")
	if err != nil {
		return nil, err
	}

	migrations := []string{}
	for rows.Next() {
		var migrationName string
		rows.Scan(&migrationName)
		migrations = append(migrations, migrationName)
	}

	return migrations, nil
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

func markAsApplied(migration string, db *sql.DB) error {
	_, err := db.Exec("INSERT INTO gomigrate_meta (name) VALUES (?)", migration)
	if err != nil {
		return errors.New("failed to mark migration as applied")
	}

	return nil
}

// CREATE TABLE IF NOT EXISTS gomigrate_meta
// (
//     name VARCHAR(255) NOT NULL
//         PRIMARY KEY,
//     CONSTRAINT name
//         UNIQUE (name)
// );

// INSERT INTO gomigrate_meta (name) values (?);

// SELECT name FROM gomigrate_meta;
