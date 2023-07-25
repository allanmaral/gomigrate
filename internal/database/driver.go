package database

import (
	"fmt"
	"net/url"
	"sync"
)

var driversMu sync.RWMutex
var drivers = make(map[string]Driver)

type Driver interface {
	Url(conf *ConnectionParams) *url.URL

	Open(url string) (Driver, error)

	Close() error

	Run(migration string) error

	AppliedMigrations() ([]string, error)

	MarkAsApplied(migration string) error

	RemoveApplied(migration string) error
}

func Url(conf *ConnectionParams) (*url.URL, error) {
	provider := conf.Provider

	driversMu.RLock()
	d, ok := drivers[provider]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("database driver: unknown driver %v", provider)
	}

	return d.Url(conf), nil
}

func Open(rawUrl string) (Driver, error) {
	purl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	provider := purl.Scheme

	driversMu.RLock()
	d, ok := drivers[provider]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("database driver: unknown driver %v", provider)
	}

	return d.Open(rawUrl)
}

func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("Register called twice for driver " + name)
	}
	drivers[name] = driver
}
