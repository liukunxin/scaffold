package snapshot

import "sync"

// Store keeps runtime snapshots for replay/recovery.
type Store struct {
	mu   sync.RWMutex
	data map[string]map[string]any
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]map[string]any),
	}
}

func (s *Store) Save(sessionID string, state map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[string]any, len(state))
	for k, v := range state {
		copied[k] = v
	}
	s.data[sessionID] = copied
}

func (s *Store) Load(sessionID string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.data[sessionID]
	if !ok {
		return nil
	}
	copied := make(map[string]any, len(state))
	for k, v := range state {
		copied[k] = v
	}
	return copied
}
