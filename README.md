# Multiplayer Lobby Package

A reusable, generic Go package for managing multiplayer lobbies, designed for use in any game or real-time application.

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/yourorg/multiplayer-lobby"
)

func main() {
	events := &lobby.LobbyEvents{
		OnPlayerJoin: func(lobby *lobby.Lobby, player *lobby.Player) {
			fmt.Printf("%s joined lobby %s\n", player.Username, lobby.Name)
		},
		OnLobbyFull: func(lobby *lobby.Lobby) {
			fmt.Printf("Lobby %s is now full!\n", lobby.Name)
		},
	}
	manager := lobby.NewLobbyManagerWithEvents(events)

	lobby1, _ := manager.CreateLobby("Room1", 2, true, nil)
	p1 := &lobby.Player{ID: "p1", Username: "Alice"}
	p2 := &lobby.Player{ID: "p2", Username: "Bob"}

	manager.JoinLobby(lobby1.ID, p1)
	manager.JoinLobby(lobby1.ID, p2)
	manager.LeaveLobby(lobby1.ID, p1.ID)
}
```

## Architectural Style

This package uses a **modular, layered architecture** inspired by Clean Architecture, but simplified for a Go library context:

- **Core Domain Layer:**
  - Types: `Lobby`, `Player`, `LobbyID`, `PlayerID`, `LobbyState`
  - Pure business logic and state transitions (e.g., add/remove player, ready status, state changes)
  - No dependencies on storage, networking, or game logic

- **Repository/Storage Abstraction:**
  - Interfaces for lobby storage (e.g., `LobbyRepository`)
  - In-memory implementation provided by default
  - Pluggable: users can implement their own (e.g., Redis, SQL)

- **Service/Manager Layer:**
  - Orchestrates lobby creation, joining, leaving, and state updates
  - Event/callback system for host app integration (on join, on full, etc.)

- **Extensibility Points:**
  - Custom metadata for lobbies and players
  - Hooks/callbacks for integration with game logic or external systems

- **No Transport/Networking:**
  - The package does not handle WebSockets, HTTP, or any network protocol
  - The host application wires up networking and calls the package API

- **Testing and Documentation:**
  - Unit tests for all core logic
  - Example usage in this README

### Summary Table

| Layer/Component      | Responsibility                                 |
|----------------------|------------------------------------------------|
| Domain (Types/Logic) | Lobby, Player, state transitions, rules        |
| Repository           | Storage abstraction (in-memory, pluggable)     |
| Manager/Service      | Orchestrates lobby operations, event hooks     |
| Extensibility        | Metadata, callbacks, custom logic              |
| No Networking        | Host app handles transport                     |

This style ensures the package is reusable, testable, and easy to integrate into any Go project. 