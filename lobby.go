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
	ID         LobbyID                // Unique identifier for the lobby
	Name       string                 // Human-readable lobby name
	MaxPlayers int                    // Maximum number of players allowed
	CreatedAt  time.Time              // Lobby creation timestamp
	Public     bool                   // Whether the lobby is public or private
	Players    []*Player              // List of players in the lobby
	State      LobbyState             // Current state of the lobby
	Metadata   map[string]interface{} // Custom fields for extensibility
}
