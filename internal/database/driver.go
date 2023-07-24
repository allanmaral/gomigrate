package database

import (
	"fmt"
	"sync"
)

var driversMu sync.RWMutex
var drivers = make(map[string]Driver)

type Driver interface {
	Open(provider string) (Driver, error)

	Close() error

	Run(migration string) error
}

func Open(provider string) (Driver, error) {
	driversMu.RLock()
	d, ok := drivers[provider]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("database driver: unknown driver %v", provider)
	}

	return d.Open(provider)
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
