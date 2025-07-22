# Multiplayer Lobby Package

A production-ready, reusable Go package for managing multiplayer lobbies with automatic reconnection, session management, and comprehensive event handling. Designed for games and real-time applications.

## Features

- ✅ **Automatic Reconnection**: Players automatically return to their lobby when reconnecting
- ✅ **Session Management**: Robust user session tracking with cleanup
- ✅ **Event-Driven Architecture**: Comprehensive callback system for all lobby events
- ✅ **Thread-Safe**: Full concurrency support for high-performance applications
- ✅ **Transport-Agnostic**: Works with WebSocket, HTTP, gRPC, or any transport
- ✅ **Extensible**: Custom metadata, pluggable storage, and flexible event hooks
- ✅ **Production-Ready**: Connection monitoring, timeout handling, and error recovery

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    lobby "github.com/jonosm/multiplayer-lobby"
)

func main() {
    // Create session manager for user tracking
    sessionManager := lobby.NewSessionManager()
    
    // Set up event handlers
    events := &lobby.LobbyEvents{
        OnPlayerJoin: func(lobby *lobby.Lobby, player *lobby.Player) {
            fmt.Printf("🎮 %s joined lobby %s\n", player.Username, lobby.Name)
        },
        OnPlayerLeave: func(lobby *lobby.Lobby, player *lobby.Player) {
            fmt.Printf("👋 %s left lobby %s\n", player.Username, lobby.Name)
        },
        OnLobbyFull: func(lobby *lobby.Lobby) {
            fmt.Printf("🏠 Lobby %s is now full!\n", lobby.Name)
        },
        OnLobbyEmpty: func(lobby *lobby.Lobby) {
            fmt.Printf("🗑️ Lobby %s is empty, cleaning up\n", lobby.Name)
        },
        OnLobbyDeleted: func(lobby *lobby.Lobby) {
            fmt.Printf("🗑️ Lobby %s was deleted\n", lobby.Name)
        },
    }
    
    // Create lobby manager with events
    manager := lobby.NewLobbyManagerWithEvents(events)
    
    // Create a session for a user
    session := sessionManager.CreateSession("alice")
    
    // Create a lobby
    lobby1, _ := manager.CreateLobby("Game Room", 4, true, nil, session.ID)
    
    // Join players
    p1 := &lobby.Player{ID: lobby.PlayerID(session.ID), Username: "Alice"}
    manager.JoinLobby(lobby1.ID, p1)
    
    // Track lobby membership for auto-reconnection
    sessionManager.SetLobbyID(session.ID, string(lobby1.ID))
    
    fmt.Printf("Created lobby: %s with %d players\n", lobby1.Name, len(lobby1.Players))
}
```

## Core Concepts

### Session Management

The package includes robust session management for handling user connections and reconnections:

```go
// Create session manager
sessionManager := lobby.NewSessionManager()

// Create user session
session := sessionManager.CreateSession("alice")

// Track lobby membership for auto-reconnection
sessionManager.SetLobbyID(session.ID, "lobby123")

// Handle reconnection
if existingSession, exists := sessionManager.GetSessionByUsername("alice"); exists {
    // User is reconnecting - restore their lobby membership
    if existingSession.LobbyID != "" {
        // Auto-rejoin to previous lobby
    }
}
```

### Auto-Reconnection

Players automatically return to their lobby when reconnecting:

```go
// When a user registers with an existing username
func RegisterUserHandler(deps *HandlerDeps) MessageHandler {
    return func(conn Conn, msg IncomingMessage) error {
        // Check for existing session
        if existingSession, exists := deps.SessionManager.GetSessionByUsername(req.Username); exists {
            // Force remove old session and create new one
            deps.SessionManager.ForceRemoveSession(existingSession.ID)
            session := deps.SessionManager.CreateSession(req.Username)
            
            // Restore lobby membership if user was in a lobby
            if existingSession.LobbyID != "" {
                session.LobbyID = existingSession.LobbyID
                // Auto-rejoin to lobby...
            }
        }
    }
}
```

### Event System

Comprehensive event handling for all lobby operations:

```go
events := &lobby.LobbyEvents{
    // Player events
    OnPlayerJoin: func(lobby *lobby.Lobby, player *lobby.Player) {
        // Handle player joining
    },
    OnPlayerLeave: func(lobby *lobby.Lobby, player *lobby.Player) {
        // Handle player leaving
    },
    OnPlayerReady: func(lobby *lobby.Lobby, player *lobby.Player) {
        // Handle ready status change
    },
    
    // Lobby events
    OnLobbyFull: func(lobby *lobby.Lobby) {
        // Handle lobby becoming full
    },
    OnLobbyEmpty: func(lobby *lobby.Lobby) {
        // Handle lobby becoming empty
    },
    OnLobbyDeleted: func(lobby *lobby.Lobby) {
        // Handle lobby deletion
    },
    OnLobbyStateChange: func(lobby *lobby.Lobby) {
        // Handle any lobby state change
    },
    
    // Broadcasting
    Broadcaster: func(userID string, message interface{}) {
        // Send message to specific user
    },
    
    // Custom logic
    CanStartGame: func(lobby *lobby.Lobby, userID string) bool {
        // Custom game start validation
        return lobby.OwnerID == userID
    },
}
```

## API Reference

### Core Types

```go
// Lobby represents a multiplayer lobby
type Lobby struct {
    ID         LobbyID                // Unique identifier
    Name       string                 // Human-readable name
    MaxPlayers int                    // Maximum players allowed
    CreatedAt  time.Time              // Creation timestamp
    Public     bool                   // Public or private
    Players    []*Player              // List of players
    State      LobbyState             // Current state
    Metadata   map[string]interface{} // Custom data
    OwnerID    string                 // Lobby owner
}

// Player represents a player in a lobby
type Player struct {
    ID       PlayerID               // Unique identifier
    Username string                 // Display name
    Ready    bool                   // Ready status
    Metadata map[string]interface{} // Custom data
}

// UserSession represents an active user session
type UserSession struct {
    ID       string    // Unique identifier
    Username string    // Username
    Active   bool      // Connection status
    LobbyID  string    // Current lobby (for reconnection)
    LastSeen time.Time // Last activity timestamp
}
```

### Session Manager

```go
// Create new session manager
sessionManager := lobby.NewSessionManager()

// Session operations
session := sessionManager.CreateSession("username")
session := sessionManager.CreateSessionWithID("user123", "username")
session, exists := sessionManager.GetSessionByID("user123")
session, exists := sessionManager.GetSessionByUsername("username")

// Lobby membership tracking
sessionManager.SetLobbyID("user123", "lobby456")
lobbyID, exists := sessionManager.GetLobbyID("user123")
sessionManager.ClearLobbyID("user123")

// Session lifecycle
sessionManager.RemoveSession("user123")
sessionManager.ForceRemoveSession("user123")
session, reconnected := sessionManager.ReconnectSession("username")

// Cleanup
sessionManager.CleanupStaleSessions(10 * time.Minute)
```

### Lobby Manager

```go
// Create lobby manager
manager := lobby.NewLobbyManager()
manager := lobby.NewLobbyManagerWithEvents(events)

// Lobby operations
lobby, err := manager.CreateLobby("name", 4, true, metadata, ownerID)
err := manager.JoinLobby(lobbyID, player)
err := manager.LeaveLobby(lobbyID, playerID)
err := manager.DeleteLobby(lobbyID)

// Player operations
err := manager.SetPlayerReady(lobbyID, playerID, true)
err := manager.StartGame(lobbyID, userID)

// Queries
lobby, exists := manager.GetLobbyByID(lobbyID)
lobbies := manager.ListLobbies()
```

### Message Handlers

```go
// Create message router
router := lobby.NewMessageRouter()

// Register handlers using type-safe action constants
router.Handle(lobby.ActionRegisterUser, lobby.RegisterUserHandler(deps))
router.Handle(lobby.ActionCreateLobby, lobby.CreateLobbyHandler(deps))
router.Handle(lobby.ActionJoinLobby, lobby.JoinLobbyHandler(deps))
router.Handle(lobby.ActionLeaveLobby, lobby.LeaveLobbyHandler(deps))
router.Handle(lobby.ActionSetReady, lobby.SetReadyHandler(deps))
router.Handle(lobby.ActionListLobbies, lobby.ListLobbiesHandler(deps))
router.Handle(lobby.ActionStartGame, lobby.StartGameHandler(deps, validateGameStart))
router.Handle(lobby.ActionGetLobbyInfo, lobby.GetLobbyInfoHandler(deps, responseBuilder))
router.Handle(lobby.ActionLogout, lobby.LogoutHandler(deps))

// Dispatch messages
err := router.Dispatch(conn, message)
```

## WebSocket Integration

The package includes a complete WebSocket demo showing how to integrate with real-time applications:

### Demo Server Setup

```go
package main

import (
    "github.com/gorilla/websocket"
    lobby "github.com/jonosm/multiplayer-lobby"
)

func main() {
    // Create managers
    sessionManager := lobby.NewSessionManager()
    manager := lobby.NewLobbyManagerWithEvents(&lobby.LobbyEvents{
        Broadcaster: func(userID string, message interface{}) {
            // Send message to user's WebSocket connection
        },
        OnLobbyDeleted: func(l *lobby.Lobby) {
            // Clear lobby membership when lobby is deleted
            for _, player := range l.Players {
                sessionManager.ClearLobbyID(string(player.ID))
            }
        },
    })
    
    // Set up handlers
    deps := &lobby.HandlerDeps{
        SessionManager: sessionManager,
        LobbyManager:   manager,
        ConnToUserID:   make(map[interface{}]string),
    }
    
    router := lobby.NewMessageRouter()
    router.Handle("register_user", lobby.RegisterUserHandler(deps))
    router.Handle("create_lobby", lobby.CreateLobbyHandler(deps))
    // ... register other handlers
    
    // WebSocket endpoint
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        conn, _ := upgrader.Upgrade(w, r, nil)
        
        // Set up connection monitoring
        conn.SetCloseHandler(func(code int, text string) error {
            // Clean up session on connection close
            return nil
        })
        
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                break
            }
            router.Dispatch(conn, msg)
        }
    })
}
```

### WebSocket Message Format

```json
// Register user
{"action": "register_user", "data": {"username": "alice"}}

// Create lobby
{"action": "create_lobby", "data": {
    "name": "Game Room",
    "max_players": 4,
    "public": true,
    "user_id": "user123"
}}

// Join lobby
{"action": "join_lobby", "data": {
    "lobby_id": "Game Room",
    "user_id": "user123"
}}

// Set ready status
{"action": "set_ready", "data": {
    "lobby_id": "Game Room",
    "user_id": "user123",
    "ready": true
}}

// Start game
{"action": "start_game", "data": {
    "lobby_id": "Game Room",
    "user_id": "user123"
}}
```

## Testing

### Unit Tests

```bash
cd multiplayer-lobby
go test ./...
```

### Integration Testing

The package includes comprehensive integration tests:

```bash
cd lobby-demo
npm install -g wscat  # For WebSocket testing
cd server && go run main.go
```

Then test with wscat:
```bash
wscat -c ws://localhost:8080/ws
```

### Auto-Reconnection Test

```javascript
// Test auto-reconnection functionality
const WebSocket = require('ws');

async function testAutoReconnection() {
    // Connect and register
    const ws1 = new WebSocket('ws://localhost:8080/ws');
    ws1.send(JSON.stringify({
        action: 'register_user',
        data: { username: 'alice' }
    }));
    
    // Create lobby
    ws1.send(JSON.stringify({
        action: 'create_lobby',
        data: {
            name: 'Test Lobby',
            max_players: 4,
            public: true,
            user_id: 'user123'
        }
    }));
    
    // Close connection
    ws1.close();
    
    // Reconnect with same username
    const ws2 = new WebSocket('ws://localhost:8080/ws');
    ws2.send(JSON.stringify({
        action: 'register_user',
        data: { username: 'alice' }
    }));
    
    // Should automatically rejoin lobby
}
```

## Architecture

The package follows a clean, layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  (WebSocket Server, HTTP API, Game Engine, etc.)           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    Service Layer                            │
│  (SessionManager, LobbyManager, Event System)              │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    Domain Layer                             │
│  (Lobby, Player, Session, Business Logic)                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    Repository Layer                         │
│  (In-Memory Storage, Pluggable Backends)                   │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

- **Transport-Agnostic**: No networking code in the library
- **Event-Driven**: Comprehensive callback system for integration
- **Thread-Safe**: Full concurrency support
- **Extensible**: Custom metadata, pluggable storage, flexible events
- **Production-Ready**: Connection monitoring, error handling, cleanup

## Contributing

Contributions are welcome! The package is designed to be easily extensible. Please see the [contributing guidelines](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 