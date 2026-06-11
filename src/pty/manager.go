package pty

import (
	"fmt"
	"sync"
)

// Manager manages multiple PTY terminals by UUID.
type Manager struct {
	mu     sync.Mutex
	terms  map[string]*Terminal
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		terms: make(map[string]*Terminal),
	}
}

// Create creates a new PTY terminal with the given UUID and default size.
func (m *Manager) Create(id string) (*Terminal, error) {
	return m.CreateWithSize(id, 80, 24)
}

// CreateWithSize creates a new PTY terminal with the given UUID and dimensions.
func (m *Manager) CreateWithSize(id string, cols, rows int) (*Terminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.terms[id]; exists {
		return nil, fmt.Errorf("terminal %s already exists", id)
	}

	t, err := NewWithSize(id, cols, rows)
	if err != nil {
		return nil, err
	}

	m.terms[id] = t
	return t, nil
}

// Get returns the terminal with the given UUID.
func (m *Manager) Get(id string) *Terminal {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.terms[id]
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

// Count returns the number of active terminals.
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.terms)
}
