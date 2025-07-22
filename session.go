package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// UserSession represents an active user session (transport-agnostic)
type UserSession struct {
	ID       string
	Username string
	Active   bool      // Whether the session is currently connected
	LobbyID  string    // The lobby ID the user was last in (empty if not in a lobby)
	LastSeen time.Time // When the session was last active
	// Consumers can associate connection objects as needed
}

// SessionManager manages active user sessions (transport-agnostic)
type SessionManager struct {
	mu           sync.RWMutex
	sessions     map[string]*UserSession // userID -> session
	usernameToID map[string]string       // username -> userID

	// Session event hooks
	OnSessionCreated     func(session *UserSession)
	OnSessionReconnected func(session *UserSession)
	OnSessionRemoved     func(session *UserSession)
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*UserSession),
		usernameToID: make(map[string]string),
	}
}

// GenerateUserID creates a unique user ID
func (sm *SessionManager) GenerateUserID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession creates a new user session
func (sm *SessionManager) CreateSession(username string) *UserSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	userID := sm.GenerateUserID()
	session := &UserSession{
		ID:       userID,
		Username: username,
		Active:   true,
	}

	sm.sessions[userID] = session
	sm.usernameToID[username] = userID

	if sm.OnSessionCreated != nil {
		sm.OnSessionCreated(session)
	}

	return session
}

// CreateSessionWithID creates a session with a specific user ID (for reconnection)
func (sm *SessionManager) CreateSessionWithID(userID string, username string) *UserSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &UserSession{
		ID:       userID,
		Username: username,
		Active:   true,
	}

	sm.sessions[userID] = session
	sm.usernameToID[username] = userID

	if sm.OnSessionCreated != nil {
		sm.OnSessionCreated(session)
	}

	return session
}

// GetSessionByID retrieves a session by user ID
func (sm *SessionManager) GetSessionByID(userID string) (*UserSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[userID]
	if exists && session.Active {
		session.LastSeen = time.Now()
	}
	return session, exists
}

// GetSessionByUsername retrieves a session by username
func (sm *SessionManager) GetSessionByUsername(username string) (*UserSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	userID, exists := sm.usernameToID[username]
	if !exists {
		return nil, false
	}
	session, exists := sm.sessions[userID]
	if exists && session.Active {
		session.LastSeen = time.Now()
	}
	return session, exists
}

// RemoveSession marks a session as inactive
func (sm *SessionManager) RemoveSession(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		if session.Active {
			session.Active = false
			if sm.OnSessionRemoved != nil {
				sm.OnSessionRemoved(session)
			}
		}
	}
}

// ForceRemoveSession forcefully removes a session regardless of its state
func (sm *SessionManager) ForceRemoveSession(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.Active = false
		if sm.OnSessionRemoved != nil {
			sm.OnSessionRemoved(session)
		}
	}
}

// IsUsernameTaken checks if a username is already in use (active session)
func (sm *SessionManager) IsUsernameTaken(username string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	userID, exists := sm.usernameToID[username]
	if !exists {
		return false
	}
	session, exists := sm.sessions[userID]
	return exists && session.Active
}

// ReconnectSession reactivates a session for a username
func (sm *SessionManager) ReconnectSession(username string) (*UserSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	userID, exists := sm.usernameToID[username]
	if !exists {
		return nil, false
	}
	session, exists := sm.sessions[userID]
	if !exists || session.Active {
		return nil, false
	}
	session.Active = true
	session.LastSeen = time.Now()
	if sm.OnSessionReconnected != nil {
		sm.OnSessionReconnected(session)
	}
	return session, true
}

// SetLobbyID sets the lobby ID for a user session
func (sm *SessionManager) SetLobbyID(userID string, lobbyID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[userID]; exists {
		session.LobbyID = lobbyID
	}
}

// GetLobbyID gets the lobby ID for a user session
func (sm *SessionManager) GetLobbyID(userID string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if session, exists := sm.sessions[userID]; exists {
		return session.LobbyID, true
	}
	return "", false
}

// ClearLobbyID clears the lobby ID for a user session
func (sm *SessionManager) ClearLobbyID(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[userID]; exists {
		session.LobbyID = ""
	}
}

// CleanupStaleSessions removes sessions that have been inactive for too long
func (sm *SessionManager) CleanupStaleSessions(maxAge time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for userID, session := range sm.sessions {
		if !session.Active && now.Sub(session.LastSeen) > maxAge {
			delete(sm.sessions, userID)
			delete(sm.usernameToID, session.Username)
		}
	}
}

// API usage:
//   sm := NewSessionManager()
//   session := sm.CreateSession("alice")
//   found, ok := sm.GetSessionByID(session.ID)
//   taken := sm.IsUsernameTaken("alice")
//   sm.RemoveSession(session.ID)
