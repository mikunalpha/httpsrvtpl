package mock

import (
	"sync"

	"github.com/mikunalpha/httpsrvtpl/store"
)

// New returns a new mock.Store.
func New() store.Store {
	s := &Store{
		connected: true,
	}
	return s
}

// Store is a mock implementation of store.Store.
type Store struct {
	mu        sync.Mutex
	connected bool

	// Other fields ...
}

// Ping is a implementation of func store.Store.Ping. Try to touch the connected database.
func (s *Store) Ping() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil
	}
	return store.ErrConnectionFailed
}

// Close is a implementation of func store.Store.Close.
func (s *Store) Close() {
	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()
}
