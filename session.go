package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// UserSession represents an active user session.
// It is transport-agnostic and can be associated with any connection type.
type UserSession struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Token    string    `json:"token"`
	Active   bool      `json:"active"`
	LobbyID  string    `json:"lobby_id"`
	LastSeen time.Time `json:"last_seen"`
}

// SessionManager manages active user sessions in a thread-safe manner.
type SessionManager struct {
	mu                   sync.RWMutex
	sessions             map[string]*UserSession
	usernameToID         map[string]string
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

// GenerateSecureToken creates a cryptographically secure session token
func (sm *SessionManager) GenerateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession creates a new user session
func (sm *SessionManager) CreateSession(username string) *UserSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	userID := sm.GenerateUserID()
	token := sm.GenerateSecureToken()
	session := &UserSession{
		ID:       userID,
		Username: username,
		Token:    token,
		Active:   true,
		LastSeen: time.Now(),
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

	token := sm.GenerateSecureToken()
	session := &UserSession{
		ID:       userID,
		Username: username,
		Token:    token,
		Active:   true,
		LastSeen: time.Now(),
	}

	sm.sessions[userID] = session
	sm.usernameToID[username] = userID

	if sm.OnSessionCreated != nil {
		sm.OnSessionCreated(session)
	}

	return session
}

// ValidateSessionToken validates a session token for a given username
func (sm *SessionManager) ValidateSessionToken(username string, token string) (*UserSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	userID, exists := sm.usernameToID[username]
	if !exists {
		return nil, false
	}

	session, exists := sm.sessions[userID]
	if !exists || !session.Active {
		return nil, false
	}

	if session.Token != token {
		return nil, false
	}

	session.LastSeen = time.Now()
	return session, true
}

// ReconnectSession allows a user to reconnect with a valid token, even if their session was inactive
func (sm *SessionManager) ReconnectSession(username string, token string) (*UserSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	userID, exists := sm.usernameToID[username]
	if !exists {
		return nil, false
	}

	session, exists := sm.sessions[userID]
	if !exists {
		return nil, false
	}

	if session.Token != token {
		return nil, false
	}

	session.Active = true
	session.LastSeen = time.Now()

	if sm.OnSessionReconnected != nil {
		sm.OnSessionReconnected(session)
	}

	return session, true
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
