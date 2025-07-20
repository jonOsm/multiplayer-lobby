# WebSocket Testing Guide

This guide shows how to manually test the multiplayer lobby WebSocket API using `wscat`.

## Prerequisites

### Install wscat
```bash
npm install -g wscat
```

### Start the Backend Server
```bash
cd lobby-demo/server
go run main.go
```

The server will start on `ws://localhost:8080/ws`

## Manual Test Commands

### 1. Connect to WebSocket
```bash
wscat -c ws://localhost:8080/ws
```

### 2. Create a Lobby
```json
{"action": "create_lobby", "name": "TestLobby", "max_players": 4, "public": true}
```

**Expected Response:**
```json
{"action": "lobby_created", "lobby_id": "TestLobby", "players": [], "state": "waiting", "metadata": null}
```

### 3. List All Lobbies
```json
{"action": "list_lobbies"}
```

**Expected Response:**
```json
{"action": "lobby_list", "lobbies": ["TestLobby"]}
```

### 4. Get Lobby Information (without joining)
```json
{"action": "get_lobby_info", "lobby_id": "TestLobby"}
```

**Expected Response:**
```json
{"action": "lobby_info", "lobby_id": "TestLobby", "name": "TestLobby", "players": [], "state": "waiting", "max_players": 4, "public": true}
```

### 5. Join a Lobby
```json
{"action": "join_lobby", "lobby_id": "TestLobby", "username": "Alice"}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": false}], "state": "waiting", "metadata": null}
```

### 6. Set Player Ready Status
```json
{"action": "set_ready", "lobby_id": "TestLobby", "username": "Alice", "ready": true}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": true}], "state": "waiting", "metadata": null}
```

### 7. Join with Another Player
```json
{"action": "join_lobby", "lobby_id": "TestLobby", "username": "Bob"}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": true}, {"username": "Bob", "ready": false}], "state": "waiting", "metadata": null}
```

### 8. Set Second Player Ready
```json
{"action": "set_ready", "lobby_id": "TestLobby", "username": "Bob", "ready": true}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": true}, {"username": "Bob", "ready": true}], "state": "waiting", "metadata": null}
```

### 9. Start the Game
```json
{"action": "start_game", "lobby_id": "TestLobby", "username": "Alice"}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": true}, {"username": "Bob", "ready": true}], "state": "in_game", "metadata": null}
```

### 10. Get Lobby Info After Game Started
```json
{"action": "get_lobby_info", "lobby_id": "TestLobby"}
```

**Expected Response:**
```json
{"action": "lobby_info", "lobby_id": "TestLobby", "name": "TestLobby", "players": [{"username": "Alice", "ready": true}, {"username": "Bob", "ready": true}], "state": "in_game", "max_players": 4, "public": true}
```

### 11. Leave a Lobby
```json
{"action": "leave_lobby", "lobby_id": "TestLobby", "username": "Bob"}
```

**Expected Response:**
```json
{"action": "lobby_state", "lobby_id": "TestLobby", "players": [{"username": "Alice", "ready": true}], "state": "in_game", "metadata": null}
```

## Complete Test Session

Here's a complete test session you can copy-paste into wscat:

```bash
# Start wscat
wscat -c ws://localhost:8080/ws

# Then send these messages one by one:

{"action": "create_lobby", "name": "GameRoom", "max_players": 3, "public": true}
{"action": "list_lobbies"}
{"action": "get_lobby_info", "lobby_id": "GameRoom"}
{"action": "join_lobby", "lobby_id": "GameRoom", "username": "Player1"}
{"action": "set_ready", "lobby_id": "GameRoom", "username": "Player1", "ready": true}
{"action": "join_lobby", "lobby_id": "GameRoom", "username": "Player2"}
{"action": "set_ready", "lobby_id": "GameRoom", "username": "Player2", "ready": true}
{"action": "start_game", "lobby_id": "GameRoom", "username": "Player1"}
{"action": "get_lobby_info", "lobby_id": "GameRoom"}
{"action": "leave_lobby", "lobby_id": "GameRoom", "username": "Player1"}
```

## Error Testing

### Test Invalid Actions
```json
{"action": "invalid_action"}
```
**Expected Response:**
```json
{"action": "error", "message": "unknown action"}
```

### Test Joining Non-existent Lobby
```json
{"action": "join_lobby", "lobby_id": "NonExistent", "username": "Alice"}
```
**Expected Response:**
```json
{"action": "error", "message": "lobby does not exist"}
```

### Test Joining Full Lobby
```json
{"action": "create_lobby", "name": "FullLobby", "max_players": 1, "public": true}
{"action": "join_lobby", "lobby_id": "FullLobby", "username": "Player1"}
{"action": "join_lobby", "lobby_id": "FullLobby", "username": "Player2"}
```
**Expected Response for second join:**
```json
{"action": "error", "message": "lobby is full"}
```

### Test Starting Game with Insufficient Players
```json
{"action": "create_lobby", "name": "SoloLobby", "max_players": 4, "public": true}
{"action": "join_lobby", "lobby_id": "SoloLobby", "username": "SoloPlayer"}
{"action": "set_ready", "lobby_id": "SoloLobby", "username": "SoloPlayer", "ready": true}
{"action": "start_game", "lobby_id": "SoloLobby", "username": "SoloPlayer"}
```
**Expected Response:**
```json
{"action": "error", "message": "need at least 2 players to start the game"}
```

### Test Starting Game with Unready Players
```json
{"action": "create_lobby", "name": "UnreadyLobby", "max_players": 4, "public": true}
{"action": "join_lobby", "lobby_id": "UnreadyLobby", "username": "Player1"}
{"action": "join_lobby", "lobby_id": "UnreadyLobby", "username": "Player2"}
{"action": "set_ready", "lobby_id": "UnreadyLobby", "username": "Player1", "ready": true}
{"action": "start_game", "lobby_id": "UnreadyLobby", "username": "Player1"}
```
**Expected Response:**
```json
{"action": "error", "message": "all players must be ready to start the game"}
```

### Test Starting Game from Non-existent Lobby
```json
{"action": "start_game", "lobby_id": "NonExistent", "username": "Alice"}
```
**Expected Response:**
```json
{"action": "error", "message": "lobby not found"}
```

### Test Getting Info from Non-existent Lobby
```json
{"action": "get_lobby_info", "lobby_id": "NonExistent"}
```
**Expected Response:**
```json
{"action": "error", "message": "lobby not found"}
```

## Troubleshooting

### Connection Issues
- Ensure the backend server is running (`go run main.go`)
- Check that the WebSocket URL is correct (`ws://localhost:8080/ws`)
- Verify no firewall is blocking port 8080

### JSON Format Issues
- Ensure all JSON is properly formatted
- Check that all required fields are present
- Verify string values are quoted

### Server Errors
- Check the server console for error messages
- Restart the server if needed
- Verify all dependencies are installed 