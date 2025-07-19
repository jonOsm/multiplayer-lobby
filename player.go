package lobby

// PlayerID uniquely identifies a player.
type PlayerID string

// Player represents a player in a lobby.
type Player struct {
	ID       PlayerID               // Unique identifier for the player
	Username string                 // Player's display name
	Ready    bool                   // Whether the player is ready
	Metadata map[string]interface{} // Custom fields for extensibility
}
