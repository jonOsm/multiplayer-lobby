package lobby

// LobbyEvents holds callbacks for lobby events.
// Register functions for the events you want to handle; leave others nil.
type LobbyEvents struct {
	// OnPlayerJoin is called when a player joins a lobby.
	OnPlayerJoin func(lobby *Lobby, player *Player)
	// OnPlayerLeave is called when a player leaves a lobby.
	OnPlayerLeave func(lobby *Lobby, player *Player)
	// OnLobbyFull is called when a lobby reaches max players.
	OnLobbyFull func(lobby *Lobby)
	// OnLobbyEmpty is called when a lobby becomes empty.
	OnLobbyEmpty func(lobby *Lobby)
	// Add more hooks as needed (e.g., OnAllReady, OnStateChange)
}

// Example usage:
//   events := &LobbyEvents{
//     OnPlayerJoin: func(lobby, player) { ... },
//     OnLobbyFull: func(lobby) { ... },
//   }
//   manager := NewLobbyManagerWithEvents(events)
