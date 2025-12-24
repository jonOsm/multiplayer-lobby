package lobby

import "time"

// LobbyID uniquely identifies a lobby.
type LobbyID string

// LobbyState represents the state of a lobby.
type LobbyState int

const (
	// LobbyWaiting indicates the lobby is waiting for players.
	LobbyWaiting LobbyState = iota
	// LobbyInGame indicates the lobby is in-game.
	LobbyInGame
	// LobbyFinished indicates the lobby has finished.
	LobbyFinished
)

// Lobby represents a multiplayer lobby.
type Lobby struct {
	ID         LobbyID
	Name       string
	MaxPlayers int
	CreatedAt  time.Time
	Public     bool
	Players    []*Player
	State      LobbyState
	Metadata   map[string]interface{}
	OwnerID    string
}
