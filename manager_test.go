package lobby

import (
	"testing"
)

func TestLobbyManager_BasicFlow(t *testing.T) {
	manager := NewLobbyManager()

	// Create a lobby
	lobby, err := manager.CreateLobby("Test Lobby", 4, true, nil, "owner1")
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	// Create players
	p1 := &Player{ID: "player1", Username: "Alice"}
	p2 := &Player{ID: "player2", Username: "Bob"}

	// Join players to lobby
	if err := manager.JoinLobby(lobby.ID, p1); err != nil {
		t.Errorf("JoinLobby failed for p1: %v", err)
	}

	if err := manager.JoinLobby(lobby.ID, p2); err != nil {
		t.Errorf("JoinLobby failed for p2: %v", err)
	}

	// Verify lobby state
	if len(lobby.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(lobby.Players))
	}

	// Leave lobby
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err != nil {
		t.Errorf("LeaveLobby failed: %v", err)
	}

	if len(lobby.Players) != 1 {
		t.Errorf("Expected 1 player after leave, got %d", len(lobby.Players))
	}
}

func TestLobbyManager_Events(t *testing.T) {
	events := &LobbyEvents{
		OnPlayerJoin: func(l *Lobby, p *Player) {
			// Event handler
		},
		OnPlayerLeave: func(l *Lobby, p *Player) {
			// Event handler
		},
	}

	manager := NewLobbyManagerWithEvents(events)

	// Create a lobby
	lobby, err := manager.CreateLobby("Test Lobby", 2, true, nil, "owner1")
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	// Create and join a player
	p1 := &Player{ID: "player1", Username: "Alice"}
	if err := manager.JoinLobby(lobby.ID, p1); err != nil {
		t.Errorf("JoinLobby failed: %v", err)
	}

	// Leave lobby
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err != nil {
		t.Errorf("LeaveLobby failed: %v", err)
	}
}

func TestLobbyManager_LeaveLobbyTwice(t *testing.T) {
	manager := NewLobbyManager()

	// Create a lobby
	lobby, err := manager.CreateLobby("Test Lobby", 4, true, nil, "owner1")
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	// Create and join a player
	p1 := &Player{ID: "player1", Username: "Alice"}
	if err := manager.JoinLobby(lobby.ID, p1); err != nil {
		t.Errorf("JoinLobby failed: %v", err)
	}

	// Leave lobby first time
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err != nil {
		t.Errorf("First LeaveLobby failed: %v", err)
	}

	// Leave lobby second time (should fail)
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err == nil {
		t.Error("Second LeaveLobby should have failed")
	}
}

func TestLobbyManager_LeaveLobbyNonExistentPlayer(t *testing.T) {
	manager := NewLobbyManager()

	// Create a lobby
	lobby, err := manager.CreateLobby("Test Lobby", 4, true, nil, "owner1")
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	// Try to leave with non-existent player
	p1 := &Player{ID: "player1", Username: "Alice"}
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err == nil {
		t.Error("LeaveLobby should have failed for non-existent player")
	}
}

func TestLobbyManager_LobbyDeletionOnEmpty(t *testing.T) {
	manager := NewLobbyManager()

	// Create a lobby
	lobby, err := manager.CreateLobby("Test Lobby", 4, true, nil, "owner1")
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	// Create and join a player
	p1 := &Player{ID: "player1", Username: "Alice"}
	if err := manager.JoinLobby(lobby.ID, p1); err != nil {
		t.Errorf("JoinLobby failed for p1: %v", err)
	}

	// Leave lobby
	if err := manager.LeaveLobby(lobby.ID, p1.ID); err != nil {
		t.Errorf("LeaveLobby failed for p1: %v", err)
	}

	// Verify lobby is deleted
	if _, exists := manager.GetLobbyByID(lobby.ID); exists {
		t.Error("Lobby should be deleted when it becomes empty")
	}

	// Verify lobby is not in the list
	lobbies := manager.ListLobbies()
	for _, l := range lobbies {
		if l.ID == lobby.ID {
			t.Error("Lobby should not be in the list after deletion")
		}
	}
}

func TestSessionTokenSecurity(t *testing.T) {
	sm := NewSessionManager()

	// Create a session for "alice"
	session := sm.CreateSession("alice")
	if session == nil {
		t.Fatal("Failed to create session")
	}

	// Test 1: Valid token should work
	validSession, valid := sm.ValidateSessionToken("alice", session.Token)
	if !valid || validSession == nil {
		t.Fatal("Valid token validation failed")
	}
	if validSession.ID != session.ID {
		t.Fatal("Valid token returned wrong session")
	}

	// Test 2: Invalid token should fail
	invalidSession, valid := sm.ValidateSessionToken("alice", "invalid_token")
	if valid || invalidSession != nil {
		t.Fatal("Invalid token validation should have failed")
	}

	// Test 3: Wrong username should fail
	wrongUserSession, valid := sm.ValidateSessionToken("bob", session.Token)
	if valid || wrongUserSession != nil {
		t.Fatal("Wrong username validation should have failed")
	}

	// Test 4: Non-existent username should fail
	nonExistentSession, valid := sm.ValidateSessionToken("nonexistent", "any_token")
	if valid || nonExistentSession != nil {
		t.Fatal("Non-existent username validation should have failed")
	}

	// Test 5: Inactive session should fail
	sm.RemoveSession(session.ID)
	inactiveSession, valid := sm.ValidateSessionToken("alice", session.Token)
	if valid || inactiveSession != nil {
		t.Fatal("Inactive session validation should have failed")
	}
}

func TestSessionTokenUniqueness(t *testing.T) {
	sm := NewSessionManager()

	// Create multiple sessions
	session1 := sm.CreateSession("alice")
	session2 := sm.CreateSession("bob")
	session3 := sm.CreateSession("charlie")

	// Verify all tokens are unique
	tokens := map[string]bool{
		session1.Token: true,
		session2.Token: true,
		session3.Token: true,
	}

	if len(tokens) != 3 {
		t.Fatal("Session tokens are not unique")
	}

	// Verify each token only works for its own session
	_, ok1 := sm.ValidateSessionToken("alice", session1.Token)
	_, ok2 := sm.ValidateSessionToken("bob", session2.Token)
	_, ok3 := sm.ValidateSessionToken("charlie", session3.Token)

	if !ok1 || !ok2 || !ok3 {
		t.Fatal("Valid tokens failed validation")
	}

	// Verify cross-token validation fails
	_, ok1 = sm.ValidateSessionToken("alice", session2.Token)
	_, ok2 = sm.ValidateSessionToken("bob", session3.Token)
	_, ok3 = sm.ValidateSessionToken("charlie", session1.Token)

	if ok1 || ok2 || ok3 {
		t.Fatal("Cross-token validation should have failed")
	}
}

func TestSessionHijackingPrevention(t *testing.T) {
	sm := NewSessionManager()

	// Create a legitimate session for "alice"
	aliceSession := sm.CreateSession("alice")
	if aliceSession == nil {
		t.Fatal("Failed to create session for alice")
	}

	// Create a legitimate session for "bob"
	bobSession := sm.CreateSession("bob")
	if bobSession == nil {
		t.Fatal("Failed to create session for bob")
	}

	// Test 1: Attacker tries to claim alice's session with wrong token
	// This should FAIL (vulnerability fixed)
	attackerSession, valid := sm.ValidateSessionToken("alice", "fake_token")
	if valid || attackerSession != nil {
		t.Fatal("❌ SECURITY VULNERABILITY: Attacker was able to claim alice's session with fake token")
	}

	// Test 2: Attacker tries to claim alice's session with bob's token
	// This should FAIL (vulnerability fixed)
	attackerSession, valid = sm.ValidateSessionToken("alice", bobSession.Token)
	if valid || attackerSession != nil {
		t.Fatal("❌ SECURITY VULNERABILITY: Attacker was able to claim alice's session with bob's token")
	}

	// Test 3: Attacker tries to claim bob's session with alice's token
	// This should FAIL (vulnerability fixed)
	attackerSession, valid = sm.ValidateSessionToken("bob", aliceSession.Token)
	if valid || attackerSession != nil {
		t.Fatal("❌ SECURITY VULNERABILITY: Attacker was able to claim bob's session with alice's token")
	}

	// Test 4: Legitimate user with correct token should work
	// This should PASS (legitimate use case)
	legitimateSession, valid := sm.ValidateSessionToken("alice", aliceSession.Token)
	if !valid || legitimateSession == nil {
		t.Fatal("❌ Legitimate user with correct token was rejected")
	}
	if legitimateSession.ID != aliceSession.ID {
		t.Fatal("❌ Legitimate user got wrong session")
	}

	// Test 5: Attacker tries to claim non-existent user's session
	// This should FAIL
	attackerSession, valid = sm.ValidateSessionToken("nonexistent", "any_token")
	if valid || attackerSession != nil {
		t.Fatal("❌ SECURITY VULNERABILITY: Attacker was able to claim non-existent user's session")
	}

	t.Log("✅ All session hijacking attempts were properly blocked")
}
