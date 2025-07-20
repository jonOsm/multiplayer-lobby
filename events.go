package lobby

// Broadcaster is a function that sends a message to a user by userID.
type Broadcaster func(userID string, message interface{})

// LobbyEvents holds callbacks for lobby events.
// Register functions for the events you want to handle; leave others nil.
type LobbyEvents struct {
	// Specific events
	OnPlayerJoin  func(lobby *Lobby, player *Player) // Called when a player joins a lobby
	OnPlayerLeave func(lobby *Lobby, player *Player) // Called when a player leaves a lobby
	OnPlayerReady func(lobby *Lobby, player *Player) // Called when a player toggles ready status
	OnLobbyFull   func(lobby *Lobby)                 // Called when a lobby reaches max players
	OnLobbyEmpty  func(lobby *Lobby)                 // Called when a lobby becomes empty

	// Broad event
	OnLobbyStateChange func(lobby *Lobby) // Called on any lobby state change (join, leave, ready, etc.)

	// Broadcasting
	Broadcaster Broadcaster // Optional: set by the application

	// Message builder for lobby state broadcasts
	LobbyStateBuilder func(lobby *Lobby) interface{} // Optional: set by the application
}

// BroadcastToLobby sends a message to all players in the lobby using the registered Broadcaster.
func (m *LobbyManager) BroadcastToLobby(l *Lobby, message interface{}) {
	if m.Events == nil || m.Events.Broadcaster == nil {
		return
	}
	for _, player := range l.Players {
		m.Events.Broadcaster(string(player.ID), message)
	}
}

// Example usage:
//   events := &LobbyEvents{
//     OnPlayerJoin: func(lobby, player) { ... },
//     OnLobbyStateChange: func(lobby) { ... },
//   }
//   manager := NewLobbyManagerWithEvents(events)
