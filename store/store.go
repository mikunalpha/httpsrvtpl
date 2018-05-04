package store

import "errors"

var (
	// ErrNotFound indicates wanted data is not found.
	ErrNotFound = errors.New("not found")
	// ErrDuplicate indicates writing operation cause duplicate data conflict.
	ErrDuplicate = errors.New("duplicate")
	// ErrConnectionFailed indicates the connection of database is failed.
	ErrConnectionFailed = errors.New("connection failed")
)

// Store is a data store interface.
type Store interface {
	Ping() error
	Close()
}
