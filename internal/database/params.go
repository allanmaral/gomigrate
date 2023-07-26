package database

import "net/url"

type ConnectionParams struct {
	User                 string
	Password             string
	Database             string
	Hostname             string
	Port                 int32
	Provider             string
	MigrationsTable      string
	MigrationsNameColumn string
}

func RemoveCustomQuery(u *url.URL) *url.URL {
	ux := *u
	vx := make(url.Values)
	for k, v := range ux.Query() {
		if len(k) <= 1 || k[0:2] != "x-" {
			vx[k] = v
		}
	}
	ux.RawQuery = vx.Encode()
	return &ux
}
