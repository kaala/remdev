package pty

import (
	"fmt"
	"sync"
)

// Manager manages multiple PTY terminals by UUID.
type Manager struct {
	mu      sync.Mutex
	terms   map[string]*Terminal
	WorkDir string // working directory for new terminals (empty = inherit)
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		terms: make(map[string]*Terminal),
	}
}

// CreateWithSize creates a new PTY terminal with the given UUID and dimensions.
func (m *Manager) CreateWithSize(id string, cols, rows int) (*Terminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.terms[id]; exists {
		return nil, fmt.Errorf("terminal %s already exists", id)
	}

	t, err := NewWithSize(id, cols, rows, m.WorkDir)
	if err != nil {
		return nil, err
	}

	m.terms[id] = t
	return t, nil
}

// Remove removes a terminal from the manager and closes it.
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.terms[id]; ok {
		t.Close()
		delete(m.terms, id)
	}
}
