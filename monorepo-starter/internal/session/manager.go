package session

import "sync"

// Manager keeps session-scoped metadata in runtime.
type Manager struct {
	mu   sync.RWMutex
	data map[string]map[string]any
}

func NewManager() *Manager {
	return &Manager{
		data: make(map[string]map[string]any),
	}
}

func (m *Manager) Set(sessionID string, values map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make(map[string]any, len(values))
	for k, v := range values {
		copied[k] = v
	}
	m.data[sessionID] = copied
}

func (m *Manager) Get(sessionID string) map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values, ok := m.data[sessionID]
	if !ok {
		return nil
	}
	copied := make(map[string]any, len(values))
	for k, v := range values {
		copied[k] = v
	}
	return copied
}
