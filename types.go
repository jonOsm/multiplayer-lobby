package lobby

// Request and response types for message router

type RegisterUserRequest struct {
	Action   string `json:"action"`
	Username string `json:"username"`
}

type RegisterUserResponse struct {
	Action   string `json:"action"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type CreateLobbyRequest struct {
	Action     string                 `json:"action"`
	Name       string                 `json:"name"`
	MaxPlayers int                    `json:"max_players"`
	Public     bool                   `json:"public"`
	UserID     string                 `json:"user_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type JoinLobbyRequest struct {
	Action  string `json:"action"`
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
}

type LeaveLobbyRequest struct {
	Action  string `json:"action"`
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
}

type SetReadyRequest struct {
	Action  string `json:"action"`
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
	Ready   bool   `json:"ready"`
}

type ListLobbiesRequest struct {
	Action string `json:"action"`
}

type StartGameRequest struct {
	Action  string `json:"action"`
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
}

type GetLobbyInfoRequest struct {
	Action  string `json:"action"`
	LobbyID string `json:"lobby_id"`
}

type LobbyInfoResponse struct {
	Action     string        `json:"action"`
	LobbyID    string        `json:"lobby_id"`
	Name       string        `json:"name"`
	Players    []PlayerState `json:"players"`
	State      string        `json:"state"`
	MaxPlayers int           `json:"max_players"`
	Public     bool          `json:"public"`
}

type ErrorResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

type LobbyStateResponse struct {
	Action   string                 `json:"action"`
	LobbyID  string                 `json:"lobby_id"`
	Players  []PlayerState          `json:"players"`
	State    string                 `json:"state"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type PlayerState struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Ready        bool   `json:"ready"`
	CanStartGame bool   `json:"can_start_game"`
}

type LobbyListResponse struct {
	Action  string   `json:"action"`
	Lobbies []string `json:"lobbies"`
}
