package lobby

import (
	"encoding/json"
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
