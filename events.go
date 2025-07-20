package lobby

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
}

// Example usage:
//   events := &LobbyEvents{
//     OnPlayerJoin: func(lobby, player) { ... },
//     OnLobbyStateChange: func(lobby) { ... },
//   }
//   manager := NewLobbyManagerWithEvents(events)
