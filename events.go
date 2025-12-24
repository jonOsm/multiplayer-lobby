package lobby

// Broadcaster sends a message to a user by their userID.
type Broadcaster func(userID string, message interface{})

// LobbyEvents holds callbacks for lobby-related events.
type LobbyEvents struct {
	OnPlayerJoin       func(lobby *Lobby, player *Player)
	OnPlayerLeave      func(lobby *Lobby, player *Player)
	OnPlayerReady      func(lobby *Lobby, player *Player)
	OnLobbyFull        func(lobby *Lobby)
	OnLobbyEmpty       func(lobby *Lobby)
	OnLobbyDeleted     func(lobby *Lobby)
	OnLobbyStateChange func(lobby *Lobby)
	Broadcaster        Broadcaster
	LobbyStateBuilder  func(lobby *Lobby) interface{}
	CanStartGame       func(lobby *Lobby, userID string) bool
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
