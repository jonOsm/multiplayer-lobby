package lobby

import (
	"encoding/json"
	"log"
)

// HandlerDeps contains dependencies required by message handlers.
type HandlerDeps struct {
	SessionManager *SessionManager
	LobbyManager   *LobbyManager
	ConnToUserID   map[interface{}]string
}

// validateSessionToken validates a session token and returns the session if valid.
func validateSessionToken(deps *HandlerDeps, userID string, token string) (*UserSession, error) {
	session, exists := deps.SessionManager.GetSessionByID(userID)
	if !exists || !session.Active {
		return nil, ErrUserInactive(userID)
	}

	if session.Token != token {
		return nil, ErrInvalidToken("authentication")
	}

	return session, nil
}

// RegisterUserHandler handles the "register_user" action.
func RegisterUserHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req RegisterUserRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("register_user").ToErrorResponse())
		}

		if req.Token != "" {
			var existingSession *UserSession
			var valid bool

			existingSession, valid = deps.SessionManager.ValidateSessionToken(req.Username, req.Token)
			if !valid {
				existingSession, valid = deps.SessionManager.ReconnectSession(req.Username, req.Token)
			}
			
			if valid {
				log.Printf("Valid reconnection for %s with token", req.Username)

				if deps.ConnToUserID != nil {
					deps.ConnToUserID[conn] = existingSession.ID
				}

				registerResponse := RegisterUserResponse{
					Action:   "user_registered",
					UserID:   existingSession.ID,
					Username: existingSession.Username,
					Token:    existingSession.Token,
				}

				if existingSession.LobbyID != "" {
					lobby, exists := deps.LobbyManager.GetLobbyByID(LobbyID(existingSession.LobbyID))
					if exists {
						playerStillInLobby := false
						for _, p := range lobby.Players {
							if p.Username == req.Username {
								playerStillInLobby = true
								break
							}
						}

							if !playerStillInLobby {
								player := &Player{ID: PlayerID(existingSession.ID), Username: existingSession.Username}
							err := deps.LobbyManager.JoinLobby(LobbyID(existingSession.LobbyID), player)
							if err == nil {
								// Send both responses: user_registered first, then lobby_state
								if err := conn.WriteJSON(registerResponse); err != nil {
									return err
								}
								// Send lobby state response to trigger navigation back to lobby
								responseBuilder := NewResponseBuilder(deps.LobbyManager)
								lobbyState := responseBuilder.BuildLobbyStateResponse(lobby)
								return conn.WriteJSON(lobbyState)
							} else {
								deps.SessionManager.ClearLobbyID(existingSession.ID)
							}
						} else {
							// Player is still in lobby, send both responses
							if err := conn.WriteJSON(registerResponse); err != nil {
								return err
							}
							// Send lobby state response
							responseBuilder := NewResponseBuilder(deps.LobbyManager)
							lobbyState := responseBuilder.BuildLobbyStateResponse(lobby)
							return conn.WriteJSON(lobbyState)
						}
					} else {
						deps.SessionManager.ClearLobbyID(existingSession.ID)
					}
				}

				return conn.WriteJSON(registerResponse)
			} else {
				log.Printf("Invalid token for reconnection attempt by %s", req.Username)
				return conn.WriteJSON(ErrInvalidToken("register_user").ToErrorResponse())
			}
		}

		// Check if username is already taken (for new registrations)
		if deps.SessionManager.IsUsernameTaken(req.Username) {
			log.Printf("Username %s is already taken", req.Username)
			return conn.WriteJSON(ErrUsernameTaken("register_user").ToErrorResponse())
		}

		// Create new session for new user
		session := deps.SessionManager.CreateSession(req.Username)
		if deps.ConnToUserID != nil {
			deps.ConnToUserID[conn] = session.ID
		}

		response := RegisterUserResponse{
			Action:   "user_registered",
			UserID:   session.ID,
			Username: session.Username,
			Token:    session.Token,
		}

		return conn.WriteJSON(response)
	}
}

// CreateLobbyHandler handles the "create_lobby" action.
func CreateLobbyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req CreateLobbyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("create_lobby").ToErrorResponse())
		}

		session, err := validateSessionToken(deps, req.UserID, req.Token)
		if err != nil {
			if lobbyErr, ok := err.(*LobbyError); ok {
				return conn.WriteJSON(lobbyErr.ToErrorResponse())
			}
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		createdLobby, err := deps.LobbyManager.CreateLobby(req.Name, req.MaxPlayers, req.Public, req.Metadata, session.ID)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		player := &Player{ID: PlayerID(session.ID), Username: session.Username}
		err = deps.LobbyManager.JoinLobby(createdLobby.ID, player)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, "failed to join creator to lobby: "+err.Error()).ToErrorResponse())
		}

		deps.SessionManager.SetLobbyID(session.ID, string(createdLobby.ID))

		responseBuilder := NewResponseBuilder(deps.LobbyManager)
		lobbyState := responseBuilder.BuildLobbyStateResponse(createdLobby)
		return conn.WriteJSON(lobbyState)
	}
}

// JoinLobbyHandler handles the "join_lobby" action.
func JoinLobbyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req JoinLobbyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("join_lobby").ToErrorResponse())
		}

		session, err := validateSessionToken(deps, req.UserID, req.Token)
		if err != nil {
			if lobbyErr, ok := err.(*LobbyError); ok {
				return conn.WriteJSON(lobbyErr.ToErrorResponse())
			}
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		player := &Player{ID: PlayerID(session.ID), Username: session.Username}
		err = deps.LobbyManager.JoinLobby(LobbyID(req.LobbyID), player)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		deps.SessionManager.SetLobbyID(session.ID, req.LobbyID)

		lobby, exists := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if exists {
			responseBuilder := NewResponseBuilder(deps.LobbyManager)
			lobbyState := responseBuilder.BuildLobbyStateResponse(lobby)
			return conn.WriteJSON(lobbyState)
		}
		return nil
	}
}

// LeaveLobbyHandler handles the "leave_lobby" action.
func LeaveLobbyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req LeaveLobbyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("leave_lobby").ToErrorResponse())
		}

		session, err := validateSessionToken(deps, req.UserID, req.Token)
		if err != nil {
			if lobbyErr, ok := err.(*LobbyError); ok {
				return conn.WriteJSON(lobbyErr.ToErrorResponse())
			}
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		err = deps.LobbyManager.LeaveLobby(LobbyID(req.LobbyID), PlayerID(session.ID))
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		deps.SessionManager.ClearLobbyID(session.ID)

		return conn.WriteJSON(map[string]interface{}{
			"action":   "left_lobby",
			"lobby_id": req.LobbyID,
		})
	}
}

// SetReadyHandler handles the "set_ready" action.
func SetReadyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req SetReadyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("set_ready").ToErrorResponse())
		}

		session, err := validateSessionToken(deps, req.UserID, req.Token)
		if err != nil {
			if lobbyErr, ok := err.(*LobbyError); ok {
				return conn.WriteJSON(lobbyErr.ToErrorResponse())
			}
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		err = deps.LobbyManager.SetPlayerReady(LobbyID(req.LobbyID), PlayerID(session.ID), req.Ready)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		lobby, exists := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if exists {
			responseBuilder := NewResponseBuilder(deps.LobbyManager)
			lobbyState := responseBuilder.BuildLobbyStateResponse(lobby)
			return conn.WriteJSON(lobbyState)
		}
		return nil
	}
}

// ListLobbiesHandler handles the "list_lobbies" action.
func ListLobbiesHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		responseBuilder := NewResponseBuilder(deps.LobbyManager)
		return conn.WriteJSON(responseBuilder.BuildLobbyListResponse())
	}
}

// StartGameHandler handles the "start_game" action.
func StartGameHandler(deps *HandlerDeps, validateGameStart func(*Lobby, string) error) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req StartGameRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("start_game").ToErrorResponse())
		}

		session, err := validateSessionToken(deps, req.UserID, req.Token)
		if err != nil {
			if lobbyErr, ok := err.(*LobbyError); ok {
				return conn.WriteJSON(lobbyErr.ToErrorResponse())
			}
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		l, ok := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if !ok {
			return conn.WriteJSON(ErrLobbyNotFound(req.LobbyID).ToErrorResponse())
		}
		if err := validateGameStart(l, session.Username); err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeCannotStartGame, err.Error()).ToErrorResponse())
		}
		err = deps.LobbyManager.StartGame(LobbyID(req.LobbyID), session.ID)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}
		return nil
	}
}

// GetLobbyInfoHandler handles the "get_lobby_info" action.
func GetLobbyInfoHandler(deps *HandlerDeps, lobbyInfoResponseFromLobby func(*Lobby) LobbyInfoResponse) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req GetLobbyInfoRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("get_lobby_info").ToErrorResponse())
		}
		l, ok := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if !ok {
			return conn.WriteJSON(ErrLobbyNotFound(req.LobbyID).ToErrorResponse())
		}
		return conn.WriteJSON(lobbyInfoResponseFromLobby(l))
	}
}

// LogoutHandler handles the "logout" action.
func LogoutHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req struct {
			Action string `json:"action"`
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("logout").ToErrorResponse())
		}
		for _, lobby := range deps.LobbyManager.ListLobbies() {
			for _, player := range lobby.Players {
				if string(player.ID) == req.UserID {
					_ = deps.LobbyManager.LeaveLobby(lobby.ID, player.ID)
					break
				}
			}
		}

		deps.SessionManager.ClearLobbyID(req.UserID)
		deps.SessionManager.RemoveSession(req.UserID)
		return nil
	}
}
