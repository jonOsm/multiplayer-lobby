package lobby

import "fmt"

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// User-related errors
	ErrorCodeUserNotFound    ErrorCode = "USER_NOT_FOUND"
	ErrorCodeUserInactive    ErrorCode = "USER_INACTIVE"
	ErrorCodeUsernameTaken   ErrorCode = "USERNAME_TAKEN"
	ErrorCodeInvalidUsername ErrorCode = "INVALID_USERNAME"

	// Lobby-related errors
	ErrorCodeLobbyNotFound        ErrorCode = "LOBBY_NOT_FOUND"
	ErrorCodeLobbyFull            ErrorCode = "LOBBY_FULL"
	ErrorCodeLobbyNotWaiting      ErrorCode = "LOBBY_NOT_WAITING"
	ErrorCodePlayerNotInLobby     ErrorCode = "PLAYER_NOT_IN_LOBBY"
	ErrorCodePlayerAlreadyInLobby ErrorCode = "PLAYER_ALREADY_IN_LOBBY"
	ErrorCodeLobbyAlreadyExists   ErrorCode = "LOBBY_ALREADY_EXISTS"
	ErrorCodeLobbyExists          ErrorCode = "LOBBY_EXISTS"

	// Game-related errors
	ErrorCodeNotEnoughPlayers   ErrorCode = "NOT_ENOUGH_PLAYERS"
	ErrorCodeNotAllPlayersReady ErrorCode = "NOT_ALL_PLAYERS_READY"
	ErrorCodeCannotStartGame    ErrorCode = "CANNOT_START_GAME"

	// Message-related errors
	ErrorCodeInvalidMessage ErrorCode = "INVALID_MESSAGE"
	ErrorCodeUnknownAction  ErrorCode = "UNKNOWN_ACTION"
	ErrorCodeInvalidRequest ErrorCode = "INVALID_REQUEST"

	// System errors
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// LobbyError represents a structured error with code and message
type LobbyError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *LobbyError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ToErrorResponse converts the error to an ErrorResponse
func (e *LobbyError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Action:  "error",
		Code:    string(e.Code),
		Message: e.Message,
		Details: e.Details,
	}
}

// NewLobbyError creates a new structured error
func NewLobbyError(code ErrorCode, message string) *LobbyError {
	return &LobbyError{
		Code:    code,
		Message: message,
	}
}

// NewLobbyErrorWithDetails creates a new structured error with additional details
func NewLobbyErrorWithDetails(code ErrorCode, message, details string) *LobbyError {
	return &LobbyError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined error constructors
func ErrUserNotFound(userID string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeUserNotFound, "User not found", fmt.Sprintf("User ID: %s", userID))
}

func ErrUserInactive(userID string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeUserInactive, "User is inactive", fmt.Sprintf("User ID: %s", userID))
}

func ErrUsernameTaken(username string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeUsernameTaken, "Username already taken", fmt.Sprintf("Username: %s", username))
}

func ErrLobbyNotFound(lobbyID string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeLobbyNotFound, "Lobby not found", fmt.Sprintf("Lobby ID: %s", lobbyID))
}

func ErrLobbyFull(lobbyID string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeLobbyFull, "Lobby is full", fmt.Sprintf("Lobby ID: %s", lobbyID))
}

func ErrPlayerNotInLobby(playerID, lobbyID string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodePlayerNotInLobby, "Player not in lobby",
		fmt.Sprintf("Player ID: %s, Lobby ID: %s", playerID, lobbyID))
}

func ErrNotEnoughPlayers(required, actual int) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeNotEnoughPlayers, "Not enough players to start game",
		fmt.Sprintf("Required: %d, Actual: %d", required, actual))
}

func ErrNotAllPlayersReady() *LobbyError {
	return NewLobbyError(ErrorCodeNotAllPlayersReady, "All players must be ready to start the game")
}

func ErrInvalidMessage(action string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeInvalidMessage, "Invalid message format",
		fmt.Sprintf("Action: %s", action))
}

func ErrUnknownAction(action string) *LobbyError {
	return NewLobbyErrorWithDetails(ErrorCodeUnknownAction, "Unknown action",
		fmt.Sprintf("Action: %s", action))
}
