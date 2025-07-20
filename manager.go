package lobby

import (
	"errors"
	"sync"
	"time"
)

// LobbyManager manages lobbies and players in a thread-safe way.
type LobbyManager struct {
	mu      sync.Mutex
	lobbies map[LobbyID]*Lobby
	Events  *LobbyEvents // Optional event hooks
}

// NewLobbyManager creates a LobbyManager with no event hooks.
func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies: make(map[LobbyID]*Lobby),
	}
}

// NewLobbyManagerWithEvents creates a LobbyManager with event hooks.
func NewLobbyManagerWithEvents(events *LobbyEvents) *LobbyManager {
	return &LobbyManager{
		lobbies: make(map[LobbyID]*Lobby),
		Events:  events,
	}
}

// CreateLobby creates a new lobby with the given parameters.
// Returns an error if a lobby with the same ID already exists.
func (m *LobbyManager) CreateLobby(name string, maxPlayers int, public bool, metadata map[string]interface{}) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := LobbyID(name) // For now, use name as ID; can be replaced with UUID
	if _, exists := m.lobbies[id]; exists {
		return nil, errors.New("lobby already exists")
	}
	lobby := &Lobby{
		ID:         id,
		Name:       name,
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now(),
		Public:     public,
		Players:    []*Player{},
		State:      LobbyWaiting,
		Metadata:   metadata,
	}
	m.lobbies[id] = lobby
	if m.Events != nil && m.Events.OnLobbyStateChange != nil {
		m.Events.OnLobbyStateChange(lobby)
	}
	return lobby, nil
}

// JoinLobby adds a player to the lobby if there is space and triggers events.
// Returns an error if the lobby does not exist, is full, or the player is already in the lobby.
func (m *LobbyManager) JoinLobby(lobbyID LobbyID, player *Player) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobby, exists := m.lobbies[lobbyID]
	if !exists {
		return errors.New("lobby does not exist")
	}
	if len(lobby.Players) >= lobby.MaxPlayers {
		return errors.New("lobby is full")
	}
	for _, p := range lobby.Players {
		if p.ID == player.ID {
			return errors.New("player already in lobby")
		}
	}
	lobby.Players = append(lobby.Players, player)
	if m.Events != nil {
		if m.Events.OnPlayerJoin != nil {
			m.Events.OnPlayerJoin(lobby, player)
		}
		if len(lobby.Players) == lobby.MaxPlayers && m.Events.OnLobbyFull != nil {
			m.Events.OnLobbyFull(lobby)
		}
		if m.Events.OnLobbyStateChange != nil {
			m.Events.OnLobbyStateChange(lobby)
		}
	}
	return nil
}

// DeleteLobby removes a lobby from the manager.
// Returns an error if the lobby does not exist.
func (m *LobbyManager) DeleteLobby(lobbyID LobbyID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.lobbies[lobbyID]; !exists {
		return errors.New("lobby does not exist")
	}
	delete(m.lobbies, lobbyID)
	return nil
}

// LeaveLobby removes a player from the lobby and triggers events.
// Returns an error if the lobby or player does not exist.
// If the lobby becomes empty after the player leaves, it will be automatically deleted.
func (m *LobbyManager) LeaveLobby(lobbyID LobbyID, playerID PlayerID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobby, exists := m.lobbies[lobbyID]
	if !exists {
		return errors.New("lobby does not exist")
	}
	var leavingPlayer *Player
	newPlayers := make([]*Player, 0, len(lobby.Players))
	for _, p := range lobby.Players {
		if p.ID == playerID {
			leavingPlayer = p
			continue
		}
		newPlayers = append(newPlayers, p)
	}
	if leavingPlayer == nil {
		return errors.New("player not in lobby")
	}
	lobby.Players = newPlayers
	if m.Events != nil {
		if m.Events.OnPlayerLeave != nil {
			m.Events.OnPlayerLeave(lobby, leavingPlayer)
		}
		if len(lobby.Players) == 0 && m.Events.OnLobbyEmpty != nil {
			m.Events.OnLobbyEmpty(lobby)
		}
		if m.Events.OnLobbyStateChange != nil {
			m.Events.OnLobbyStateChange(lobby)
		}
	}
	// If lobby becomes empty, delete it
	if len(lobby.Players) == 0 {
		delete(m.lobbies, lobbyID)
	}
	return nil
}

// Add a method to toggle ready status and trigger events
func (m *LobbyManager) SetPlayerReady(lobbyID LobbyID, playerID PlayerID, ready bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobby, exists := m.lobbies[lobbyID]
	if !exists {
		return errors.New("lobby does not exist")
	}
	var targetPlayer *Player
	for _, p := range lobby.Players {
		if p.ID == playerID {
			targetPlayer = p
			break
		}
	}
	if targetPlayer == nil {
		return errors.New("player not in lobby")
	}
	if targetPlayer.Ready == ready {
		return nil // No change
	}
	targetPlayer.Ready = ready
	if m.Events != nil {
		if m.Events.OnPlayerReady != nil {
			m.Events.OnPlayerReady(lobby, targetPlayer)
		}
		if m.Events.OnLobbyStateChange != nil {
			m.Events.OnLobbyStateChange(lobby)
		}
	}
	return nil
}

// ListLobbies returns all lobbies managed by the LobbyManager.
func (m *LobbyManager) ListLobbies() []*Lobby {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobbies := make([]*Lobby, 0, len(m.lobbies))
	for _, l := range m.lobbies {
		lobbies = append(lobbies, l)
	}
	return lobbies
}

// GetLobbyByID returns a lobby by its ID and whether it exists.
func (m *LobbyManager) GetLobbyByID(id LobbyID) (*Lobby, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobby, exists := m.lobbies[id]
	return lobby, exists
}
