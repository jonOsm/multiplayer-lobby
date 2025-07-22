package lobby

import (
	"encoding/json"
)

// Action constants for type safety and IDE support
const (
	ActionRegisterUser = "register_user"
	ActionCreateLobby  = "create_lobby"
	ActionJoinLobby    = "join_lobby"
	ActionLeaveLobby   = "leave_lobby"
	ActionSetReady     = "set_ready"
	ActionListLobbies  = "list_lobbies"
	ActionStartGame    = "start_game"
	ActionGetLobbyInfo = "get_lobby_info"
	ActionLogout       = "logout"
)

// Conn is a minimal interface for sending JSON responses, transport-agnostic.
type Conn interface {
	WriteJSON(v interface{}) error
}

// IncomingMessage represents a parsed incoming message with an action.
type IncomingMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// MessageHandler processes a message for a connection.
type MessageHandler func(conn Conn, msg IncomingMessage) error

// Middleware wraps a MessageHandler for cross-cutting concerns.
type Middleware func(next MessageHandler) MessageHandler

// MessageRouter routes messages to handlers and supports middleware.
type MessageRouter struct {
	handlers   map[string]MessageHandler
	middleware []Middleware
}

// NewMessageRouter creates a new MessageRouter.
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]MessageHandler),
	}
}

// Handle registers a handler for an action.
func (r *MessageRouter) Handle(action string, handler MessageHandler) {
	r.handlers[action] = handler
}

// Use adds middleware to the router (applies to all handlers).
func (r *MessageRouter) Use(mw Middleware) {
	r.middleware = append(r.middleware, mw)
}

// SetupDefaultHandlers automatically registers all standard lobby handlers.
// This is the recommended way to set up the router - no manual wiring needed!
func (r *MessageRouter) SetupDefaultHandlers(deps *HandlerDeps) {
	r.Handle(ActionRegisterUser, RegisterUserHandler(deps))
	r.Handle(ActionCreateLobby, CreateLobbyHandler(deps))
	r.Handle(ActionJoinLobby, JoinLobbyHandler(deps))
	r.Handle(ActionLeaveLobby, LeaveLobbyHandler(deps))
	r.Handle(ActionSetReady, SetReadyHandler(deps))
	r.Handle(ActionListLobbies, ListLobbiesHandler(deps))
	r.Handle(ActionStartGame, StartGameHandler(deps, nil))       // Default validation
	r.Handle(ActionGetLobbyInfo, GetLobbyInfoHandler(deps, nil)) // Default response builder
	r.Handle(ActionLogout, LogoutHandler(deps))
}

// SetupDefaultHandlersWithCustom validates and sets up handlers with custom functions.
// Use this when you need custom game start validation or response building.
func (r *MessageRouter) SetupDefaultHandlersWithCustom(deps *HandlerDeps, options *HandlerOptions) {
	r.Handle(ActionRegisterUser, RegisterUserHandler(deps))
	r.Handle(ActionCreateLobby, CreateLobbyHandler(deps))
	r.Handle(ActionJoinLobby, JoinLobbyHandler(deps))
	r.Handle(ActionLeaveLobby, LeaveLobbyHandler(deps))
	r.Handle(ActionSetReady, SetReadyHandler(deps))
	r.Handle(ActionListLobbies, ListLobbiesHandler(deps))

	// Use custom validation if provided, otherwise use default
	gameStartValidator := options.GameStartValidator
	if gameStartValidator == nil {
		gameStartValidator = func(l *Lobby, username string) error { return nil }
	}
	r.Handle(ActionStartGame, StartGameHandler(deps, gameStartValidator))

	// Use custom response builder if provided, otherwise use default
	responseBuilder := options.ResponseBuilder
	if responseBuilder == nil {
		responseBuilder = NewResponseBuilder(deps.LobbyManager)
	}
	r.Handle(ActionGetLobbyInfo, GetLobbyInfoHandler(deps, func(l *Lobby) LobbyInfoResponse {
		return responseBuilder.BuildLobbyInfoResponse(l)
	}))

	r.Handle(ActionLogout, LogoutHandler(deps))
}

// HandlerOptions allows customization of specific handlers
type HandlerOptions struct {
	GameStartValidator func(*Lobby, string) error
	ResponseBuilder    *ResponseBuilder
}

// Dispatch parses and routes a raw message to the appropriate handler.
func (r *MessageRouter) Dispatch(conn Conn, rawMsg []byte) error {
	var msg IncomingMessage
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return conn.WriteJSON(ErrInvalidMessage("").ToErrorResponse())
	}
	handler, ok := r.handlers[msg.Action]
	if !ok {
		return conn.WriteJSON(ErrUnknownAction(msg.Action).ToErrorResponse())
	}
	// Apply middleware chain
	finalHandler := handler
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}
	return finalHandler(conn, msg)
}
