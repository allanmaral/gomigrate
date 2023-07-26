package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	nurl "net/url"

	"github.com/allanmaral/gomigrate/internal/database"
	mssql "github.com/microsoft/go-mssqldb"
)

func init() {
	database.Register("sqlserver", &SQLServer{})
}

var DefaultMigrationsTable = "schema_migrations"
var DefaultMigrationsNameColumn = "name"

var (
	ErrNilConfig            = fmt.Errorf("no config")
	ErrNoDatabaseName       = fmt.Errorf("no database name")
	ErrNoSchema             = fmt.Errorf("no schema")
	ErrCreateMigrationTable = fmt.Errorf("failed to create migration table")
)

type Config struct {
	MigrationsTable      string
	MigrationsNameColumn string
	DatabaseName         string
	SchemaName           string
}

type SQLServer struct {
	conn *sql.Conn
	db   *sql.DB

	config *Config
}

func WithInstance(instance *sql.DB, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := instance.Ping(); err != nil {
		return nil, err
	}

	if config.DatabaseName == "" {
		query := `SELECT DB_NAME()`
		var databaseName string
		if err := instance.QueryRow(query).Scan(&databaseName); err != nil {
			return nil, &database.Error{OrigErr: err, Query: []byte(query)}
		}

		if len(databaseName) == 0 {
			return nil, ErrNoDatabaseName
		}

		config.DatabaseName = databaseName
	}

	if config.SchemaName == "" {
		query := `SELECT SCHEMA_NAME()`
		var schemaName string
		if err := instance.QueryRow(query).Scan(&schemaName); err != nil {
			return nil, &database.Error{OrigErr: err, Query: []byte(query)}
		}

		if len(schemaName) == 0 {
			return nil, ErrNoSchema
		}

		config.SchemaName = schemaName
	}

	if len(config.MigrationsTable) == 0 {
		config.MigrationsTable = DefaultMigrationsTable
	}

	if len(config.MigrationsNameColumn) == 0 {
		config.MigrationsNameColumn = DefaultMigrationsNameColumn
	}

	conn, err := instance.Conn(context.Background())

	if err != nil {
		return nil, err
	}

	ss := &SQLServer{
		conn:   conn,
		db:     instance,
		config: config,
	}

	if err := ss.ensureMigrationsTable(); err != nil {
		return nil, err
	}

	return ss, nil
}

func (ss *SQLServer) Url(conf *database.ConnectionParams) *nurl.URL {
	user := conf.User
	password := conf.Password
	hostname := conf.Hostname
	port := conf.Port
	database := conf.Database
	tableName := conf.MigrationsTable
	nameColumn := conf.MigrationsNameColumn

	if user == "" {
		user = "sa"
	}
	if port == 0 {
		port = 1433
	}
	if database == "" {
		database = "master"
	}

	query := nurl.Values{}
	query.Add("database", database)
	query.Add("x-migrations-table", tableName)
	query.Add("x-name-column", nameColumn)

	return &nurl.URL{
		Scheme:   "sqlserver",
		User:     nurl.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%d", hostname, port),
		RawQuery: query.Encode(),
	}
}

func (ss *SQLServer) Open(url string) (database.Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	migrationsTable := purl.Query().Get("x-migrations-table")
	nameColumn := purl.Query().Get("x-name-column")

	filteredUrl := database.RemoveCustomQuery(purl)

	db, err := sql.Open("sqlserver", filteredUrl.String())
	if err != nil {
		return nil, err
	}

	driver, err := WithInstance(db, &Config{
		DatabaseName:         purl.Path,
		MigrationsTable:      migrationsTable,
		MigrationsNameColumn: nameColumn,
	})

	if err != nil {
		return nil, err
	}

	return driver, nil
}

func (ss *SQLServer) Close() error {
	connErr := ss.conn.Close()
	dbErr := ss.db.Close()
	if connErr != nil || dbErr != nil {
		return fmt.Errorf("conn: %v, db: %v", connErr, dbErr)
	}
	return nil
}

func (ss *SQLServer) Run(migration string) error {
	if _, err := ss.conn.ExecContext(context.Background(), migration); err != nil {
		if msErr, ok := err.(mssql.Error); ok {
			message := fmt.Sprintf("migration failed: %s", msErr.Message)
			if msErr.ProcName != "" {
				message = fmt.Sprintf("%s (proc name %s)", msErr.Message, msErr.ProcName)
			}
			return database.Error{OrigErr: err, Err: message, Query: []byte(migration), Line: uint(msErr.LineNo)}
		}
		return database.Error{OrigErr: err, Err: "migration failed", Query: []byte(migration)}
	}

	return nil
}

func (ss *SQLServer) AppliedMigrations() ([]string, error) {
	rows, err := ss.conn.QueryContext(
		context.Background(),
		`SELECT `+ss.config.MigrationsNameColumn+` FROM "`+ss.config.MigrationsTable+`" ORDER BY "`+ss.config.MigrationsNameColumn+`";`)
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

func (ss *SQLServer) MarkAsApplied(migration string) error {
	_, err := ss.conn.ExecContext(
		context.Background(),
		`INSERT INTO "`+ss.config.MigrationsTable+`" (`+ss.config.MigrationsNameColumn+`) VALUES (@p1);`,
		migration)
	if err != nil {
		return fmt.Errorf("failed to mark migration as applied")
	}

	return nil
}

func (ss *SQLServer) RemoveApplied(migration string) error {
	_, err := ss.conn.ExecContext(
		context.Background(),
		`DELETE FROM "`+ss.config.MigrationsTable+`" WHERE `+ss.config.MigrationsNameColumn+` = @p1;`,
		migration)
	if err != nil {
		return fmt.Errorf("failed to mark migration as applied")
	}

	return nil
}

func (ss *SQLServer) ensureMigrationsTable() error {
	query := `IF NOT EXISTS
		 (SELECT *
			FROM sys.tables t
								JOIN sys.schemas s ON (t.schema_id = s.schema_id)
			WHERE s.name = 'dbo'
				AND t.name = '` + ss.config.MigrationsTable + `')
		CREATE TABLE ` + ss.config.MigrationsTable + `
		(
				` + ss.config.MigrationsNameColumn + ` VARCHAR(255) NOT NULL PRIMARY KEY,
				CONSTRAINT UN__` + ss.config.MigrationsTable + `__` + ss.config.MigrationsNameColumn + ` UNIQUE (` + ss.config.MigrationsNameColumn + `)
		);`

	if _, err := ss.conn.ExecContext(context.Background(), query); err != nil {
		// return ErrCreateMigrationTable
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	return nil
}
