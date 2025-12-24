package lobby

// RegisterUserRequest represents a request to register a new user or reconnect.
type RegisterUserRequest struct {
	Username string `json:"username"`
	Token    string `json:"token,omitempty"`
}

// RegisterUserResponse represents the response after user registration.
type RegisterUserResponse struct {
	Action   string `json:"action"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// CreateLobbyRequest represents a request to create a new lobby.
type CreateLobbyRequest struct {
	Name       string                 `json:"name"`
	MaxPlayers int                    `json:"max_players"`
	Public     bool                   `json:"public"`
	UserID     string                 `json:"user_id"`
	Token      string                 `json:"token"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// JoinLobbyRequest represents a request to join an existing lobby.
type JoinLobbyRequest struct {
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
}

// LeaveLobbyRequest represents a request to leave a lobby.
type LeaveLobbyRequest struct {
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
}

// SetReadyRequest represents a request to set a player's ready status.
type SetReadyRequest struct {
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
	Ready   bool   `json:"ready"`
}

// ListLobbiesRequest represents a request to list all lobbies.
type ListLobbiesRequest struct {
	Token string `json:"token"`
}

// StartGameRequest represents a request to start a game in a lobby.
type StartGameRequest struct {
	LobbyID string `json:"lobby_id"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
}

// GetLobbyInfoRequest represents a request to get information about a lobby.
type GetLobbyInfoRequest struct {
	LobbyID string `json:"lobby_id"`
	Token   string `json:"token"`
}

// LobbyInfoResponse represents the response containing lobby information.
type LobbyInfoResponse struct {
	Action     string        `json:"action"`
	LobbyID    string        `json:"lobby_id"`
	Name       string        `json:"name"`
	Players    []PlayerState `json:"players"`
	State      string        `json:"state"`
	MaxPlayers int           `json:"max_players"`
	Public     bool          `json:"public"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Action  string `json:"action"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// LobbyStateResponse represents the current state of a lobby.
type LobbyStateResponse struct {
	Action   string                 `json:"action"`
	LobbyID  string                 `json:"lobby_id"`
	Players  []PlayerState          `json:"players"`
	State    string                 `json:"state"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PlayerState represents the state of a player in a lobby.
type PlayerState struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Ready        bool   `json:"ready"`
	CanStartGame bool   `json:"can_start_game"`
}

// LobbyListResponse represents a list of available lobbies.
type LobbyListResponse struct {
	Action  string   `json:"action"`
	Lobbies []string `json:"lobbies"`
}
