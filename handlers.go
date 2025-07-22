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
			return conn.WriteJSON(ErrorResponse{"error", "invalid register_user message"})
		}

		if deps.SessionManager.IsUsernameTaken(req.Username) {
			log.Printf("Username %s is already active, cannot reconnect", req.Username)
			return conn.WriteJSON(ErrorResponse{"error", "username already taken"})
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
			return conn.WriteJSON(ErrorResponse{"error", "invalid create_lobby message"})
		}

		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrorResponse{"error", "user not found or inactive"})
		}

		createdLobby, err := deps.LobbyManager.CreateLobby(req.Name, req.MaxPlayers, req.Public, req.Metadata, session.ID)
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}

		player := &Player{ID: PlayerID(session.ID), Username: session.Username}
		err = deps.LobbyManager.JoinLobby(createdLobby.ID, player)
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", "failed to join creator to lobby: " + err.Error()})
		}

		// Send lobby state response to confirm creation and trigger navigation
		if deps.LobbyManager.Events != nil && deps.LobbyManager.Events.LobbyStateBuilder != nil {
			lobbyState := deps.LobbyManager.Events.LobbyStateBuilder(createdLobby)
			return conn.WriteJSON(lobbyState)
		}
		return nil
	}
}

// JoinLobbyHandler handles the "join_lobby" action.
func JoinLobbyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req JoinLobbyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrorResponse{"error", "invalid join_lobby message"})
		}
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrorResponse{"error", "user not found or inactive"})
		}
		player := &Player{ID: PlayerID(session.ID), Username: session.Username}
		err := deps.LobbyManager.JoinLobby(LobbyID(req.LobbyID), player)
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}

		// Send lobby state response to confirm join and trigger navigation
		lobby, exists := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if exists && deps.LobbyManager.Events != nil && deps.LobbyManager.Events.LobbyStateBuilder != nil {
			lobbyState := deps.LobbyManager.Events.LobbyStateBuilder(lobby)
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
			return conn.WriteJSON(ErrorResponse{"error", "invalid leave_lobby message"})
		}
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrorResponse{"error", "user not found or inactive"})
		}
		err := deps.LobbyManager.LeaveLobby(LobbyID(req.LobbyID), PlayerID(session.ID))
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}
		return nil
	}
}

// SetReadyHandler handles the "set_ready" action.
func SetReadyHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req SetReadyRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrorResponse{"error", "invalid set_ready message"})
		}
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrorResponse{"error", "user not found or inactive"})
		}
		err := deps.LobbyManager.SetPlayerReady(LobbyID(req.LobbyID), PlayerID(session.ID), req.Ready)
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}

		// Send lobby state response to confirm ready status change
		lobby, exists := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if exists && deps.LobbyManager.Events != nil && deps.LobbyManager.Events.LobbyStateBuilder != nil {
			lobbyState := deps.LobbyManager.Events.LobbyStateBuilder(lobby)
			return conn.WriteJSON(lobbyState)
		}
		return nil
	}
}

// ListLobbiesHandler handles the "list_lobbies" action.
func ListLobbiesHandler(deps *HandlerDeps) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		lobbies := deps.LobbyManager.ListLobbies()
		ids := make([]string, 0, len(lobbies))
		for _, l := range lobbies {
			ids = append(ids, string(l.ID))
		}
		return conn.WriteJSON(LobbyListResponse{
			Action:  "lobby_list",
			Lobbies: ids,
		})
	}
}

// StartGameHandler handles the "start_game" action.
func StartGameHandler(deps *HandlerDeps, validateGameStart func(*Lobby, string) error) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req StartGameRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrorResponse{"error", "invalid start_game message"})
		}
		session, exists := deps.SessionManager.GetSessionByID(req.UserID)
		if !exists || !session.Active {
			return conn.WriteJSON(ErrorResponse{"error", "user not found or inactive"})
		}
		l, ok := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if !ok {
			return conn.WriteJSON(ErrorResponse{"error", "lobby not found"})
		}
		if err := validateGameStart(l, session.Username); err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}
		err := deps.LobbyManager.StartGame(LobbyID(req.LobbyID), session.ID)
		if err != nil {
			return conn.WriteJSON(ErrorResponse{"error", err.Error()})
		}
		return nil
	}
}

// GetLobbyInfoHandler handles the "get_lobby_info" action.
func GetLobbyInfoHandler(deps *HandlerDeps, lobbyInfoResponseFromLobby func(*Lobby) LobbyInfoResponse) MessageHandler {
	return func(conn Conn, msg IncomingMessage) error {
		var req GetLobbyInfoRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return conn.WriteJSON(ErrorResponse{"error", "invalid get_lobby_info message"})
		}
		l, ok := deps.LobbyManager.GetLobbyByID(LobbyID(req.LobbyID))
		if !ok {
			return conn.WriteJSON(ErrorResponse{"error", "lobby not found"})
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
			return conn.WriteJSON(ErrorResponse{"error", "invalid logout message"})
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
