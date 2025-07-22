package lobby

import (
	"encoding/json"
	"log"
)

// Handler dependencies (to be injected by the application)
type HandlerDeps struct {
	SessionManager *SessionManager
	LobbyManager   *LobbyManager
	ConnToUserID   map[interface{}]string // The application should provide a mapping from Conn to userID
}

// RegisterUserHandler handles the "register_user" action.
func RegisterUserHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req RegisterUserRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("register_user").ToErrorResponse())
		}

		if deps.SessionManager.IsUsernameTaken(req.Username) {
			log.Printf("Username %s is already active, cannot reconnect", req.Username)
			return conn.WriteJSON(ErrUsernameTaken(req.Username).ToErrorResponse())
		}

		if existingSession, exists := deps.SessionManager.GetSessionByUsername(req.Username); exists && !existingSession.Active {
			session, _ := deps.SessionManager.ReconnectSession(req.Username)
			if deps.ConnToUserID != nil {
				deps.ConnToUserID[conn] = session.ID
			}
			return conn.WriteJSON(RegisterUserResponse{
				Action:   "user_registered",
				UserID:   session.ID,
				Username: session.Username,
			})
		}

		session := deps.SessionManager.CreateSession(req.Username)
		if deps.ConnToUserID != nil {
			deps.ConnToUserID[conn] = session.ID
		}
		log.Printf("New user registered: %s with ID: %s", req.Username, session.ID)
		return conn.WriteJSON(RegisterUserResponse{
			Action:   "user_registered",
			UserID:   session.ID,
			Username: session.Username,
		})
	}
}

// CreateLobbyHandler handles the "create_lobby" action.
func CreateLobbyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req CreateLobbyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("create_lobby").ToErrorResponse())
		}

		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrUserInactive(req.UserID).ToErrorResponse())
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

		// Send lobby state response to confirm creation and trigger navigation
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
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrUserInactive(req.UserID).ToErrorResponse())
		}
		player := &Player{ID: PlayerID(session.ID), Username: session.Username}
		err := deps.LobbyManager.JoinLobby(LobbyID(req.LobbyID), player)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		// Send lobby state response to confirm join and trigger navigation
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
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrUserInactive(req.UserID).ToErrorResponse())
		}
		err := deps.LobbyManager.LeaveLobby(LobbyID(req.LobbyID), PlayerID(session.ID))
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}
		return nil
	}
}

// SetReadyHandler handles the "set_ready" action.
func SetReadyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req SetReadyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrInvalidMessage("set_ready").ToErrorResponse())
		}
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrUserInactive(req.UserID).ToErrorResponse())
		}
		err := deps.LobbyManager.SetPlayerReady(LobbyID(req.LobbyID), PlayerID(session.ID), req.Ready)
		if err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeInternalError, err.Error()).ToErrorResponse())
		}

		// Send lobby state response to confirm ready status change
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
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrUserInactive(req.UserID).ToErrorResponse())
		}
		l, ok := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if !ok {
			return conn.WriteJSON(ErrLobbyNotFound(req.LobbyID).ToErrorResponse())
		}
		if err := validateGameStart(l, session.Username); err != nil {
			return conn.WriteJSON(NewLobbyError(ErrorCodeCannotStartGame, err.Error()).ToErrorResponse())
		}
		err := deps.LobbyManager.StartGame(LobbyID(req.LobbyID), session.ID)
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
		// Remove user from any lobby they are in
		for _, lobby := range deps.LobbyManager.ListLobbies() {
			for _, player := range lobby.Players {
				if string(player.ID) == req.UserID {
					_ = deps.LobbyManager.LeaveLobby(lobby.ID, player.ID)
					break
				}
			}
		}
		// Remove the session
		deps.SessionManager.RemoveSession(req.UserID)
		return nil
	}
}
