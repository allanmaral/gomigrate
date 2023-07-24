package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/allanmaral/gomigrate/internal/config"
	"github.com/gosimple/slug"
)

var migrationTemplate = `
-- Up
BEGIN

--  Add altering commands here.
-- 
--  Example:
--  CREATE TABLE users (
--    user_id INT,
--    last_name VARCHAR(255),
--    first_name VARCHAR(255),
--    created_at TIMESTAMPTZ 
--  );

END


-- Down
BEGIN

-- Add reverting commands here.
--
-- Example:
-- DROP TABLE users;

END
`

func NewMigration(name string, c *config.Config) error {
	now := formatDate(time.Now())
	name = slug.Make(name)

	filename := fmt.Sprintf("%s-%s.sql", now, name)
	path := filepath.Join(c.MigrationsPath, filename)

	if err := createFile(path, []byte(migrationTemplate)); err != nil {
		return err
	}

	fmt.Printf("New migration was created at \"%s\".\n", path)

	return nil
}

func formatDate(date time.Time) string {
	timeStr := date.Format("2006-01-02T15:04:05")
	timeStr = strings.ReplaceAll(timeStr, "-", "")
	timeStr = strings.ReplaceAll(timeStr, "T", "")
	timeStr = strings.ReplaceAll(timeStr, ":", "")
	return timeStr
}

// TODO: Move to infra layer
func createFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return errors.New("failed write config file")
	}

	return nil
}
