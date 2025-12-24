package lobby

// PlayerID uniquely identifies a player.
type PlayerID string

// Player represents a player in a lobby.
type Player struct {
	ID       PlayerID
	Username string
	Ready    bool
	Metadata map[string]interface{}
}
