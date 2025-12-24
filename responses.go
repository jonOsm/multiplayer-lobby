package lobby

// ResponseBuilder provides standardized response formatting for the lobby system
type ResponseBuilder struct {
	manager *LobbyManager
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(manager *LobbyManager) *ResponseBuilder {
	return &ResponseBuilder{
		manager: manager,
	}
}

// BuildLobbyStateResponse creates a standardized lobby state response
func (rb *ResponseBuilder) BuildLobbyStateResponse(l *Lobby) LobbyStateResponse {
	players := make([]PlayerState, 0, len(l.Players))
	canStartGameFunc := rb.manager.Events.CanStartGame

	for _, p := range l.Players {
		canStart := false
		if canStartGameFunc != nil {
			canStart = canStartGameFunc(l, string(p.ID))
		} else {
			canStart = (l.OwnerID == string(p.ID))
		}

		players = append(players, PlayerState{
			UserID:       string(p.ID),
			Username:     p.Username,
			Ready:        p.Ready,
			CanStartGame: canStart,
		})
	}

	return LobbyStateResponse{
		Action:   "lobby_state",
		LobbyID:  string(l.ID),
		Players:  players,
		State:    lobbyStateString(l.State),
		Metadata: l.Metadata,
	}
}

// BuildLobbyInfoResponse creates a standardized lobby info response
func (rb *ResponseBuilder) BuildLobbyInfoResponse(l *Lobby) LobbyInfoResponse {
	players := make([]PlayerState, 0, len(l.Players))
	for _, p := range l.Players {
		players = append(players, PlayerState{
			UserID:       string(p.ID),
			Username:     p.Username,
			Ready:        p.Ready,
			CanStartGame: false,
		})
	}

	return LobbyInfoResponse{
		Action:     "lobby_info",
		LobbyID:    string(l.ID),
		Name:       l.Name,
		Players:    players,
		State:      lobbyStateString(l.State),
		MaxPlayers: l.MaxPlayers,
		Public:     l.Public,
	}
}

// BuildLobbyListResponse creates a standardized lobby list response
func (rb *ResponseBuilder) BuildLobbyListResponse() LobbyListResponse {
	lobbies := rb.manager.ListLobbies()
	ids := make([]string, 0, len(lobbies))
	for _, l := range lobbies {
		ids = append(ids, string(l.ID))
	}

	return LobbyListResponse{
		Action:  "lobby_list",
		Lobbies: ids,
	}
}

// BuildSuccessResponse creates a standardized success response
func (rb *ResponseBuilder) BuildSuccessResponse(action string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"action":  action,
		"success": true,
		"data":    data,
	}
}

// lobbyStateString converts lobby state to string representation
func lobbyStateString(state LobbyState) string {
	switch state {
	case LobbyWaiting:
		return "waiting"
	case LobbyInGame:
		return "in_game"
	case LobbyFinished:
		return "finished"
	default:
		return "unknown"
	}
}
