package session

import (
	"fmt"
	"mobius/pkg/agent"
)



type Manager struct {
	sessions	map[string]*Session
	activeId	string
}

type SessionInfo struct {
	ID       string
	Name     string
	IsActive bool
}


func NewManager(defaultSession *Session) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		activeId: defaultSession.ID,
	}
	// Add the default session into the map
	m.sessions[defaultSession.ID] = defaultSession
	return m
}


func (m *Manager) GetActive() (*Session, error) {
	session, exists := m.sessions[m.activeId]
	if !exists {
		return nil, fmt.Errorf("no active session found")
	}
	return session, nil // the actual pointer so memory updates persist!
}


func (m *Manager) AddSession(session *Session) {
	m.sessions[session.ID] = session
}

func (m *Manager) SwitchSession(sessionId string) (*Session, error) {
	session, exists := m.sessions[sessionId]
	if !exists {
		return nil, fmt.Errorf("session '%s' not found", sessionId)
	}
	m.activeId = session.ID //  Updates the active session pointer!
	return session, nil
}



func (m *Manager) ListSession() []SessionInfo {
	var list []SessionInfo
	for _, s := range m.sessions {
		list = append(list, SessionInfo{
			ID:       s.ID,
			Name:     s.Name,
			IsActive: s.ID == m.activeId,
		})
	}
	return list
}


// CreateSession creates and switches to a new session
func (m *Manager) CreateSession(name string, a *agent.Agent) *Session {
	sess := NewSession(a.ThreadID(), name, a)
	m.AddSession(sess)
	m.activeId = sess.ID
	return sess
}


func (m *Manager) GetUnstarted() *Session {
	for _, s := range m.sessions {
		if !s.Started {
			return s
		}
	}
	return nil
}
