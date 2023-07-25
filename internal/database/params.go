package database

type ConnectionParams struct {
	User     string
	Password string
	Database string
	Hostname string
	Port     int32
	Provider string
}
