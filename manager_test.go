package lobby

import (
	"testing"
)

func TestLobbyManager_BasicFlow(t *testing.T) {
	events := &LobbyEvents{}
	manager := NewLobbyManagerWithEvents(events)

	lobby, err := manager.CreateLobby("TestLobby", 2, true, nil)
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}
	if lobby.Name != "TestLobby" {
		t.Errorf("Expected lobby name 'TestLobby', got %s", lobby.Name)
	}

	p1 := &Player{ID: "p1", Username: "Alice"}
	p2 := &Player{ID: "p2", Username: "Bob"}

	if err := manager.JoinLobby(lobby.ID, p1); err != nil {
		t.Errorf("JoinLobby failed for p1: %v", err)
	}
	if err := manager.JoinLobby(lobby.ID, p2); err != nil {
		t.Errorf("JoinLobby failed for p2: %v", err)
	}

	lobby, _ = manager.GetLobbyByID(lobby.ID)
	if len(lobby.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(lobby.Players))
	}

	if err := manager.LeaveLobby(lobby.ID, p1.ID); err != nil {
		t.Errorf("LeaveLobby failed for p1: %v", err)
	}
	lobby, _ = manager.GetLobbyByID(lobby.ID)
	if len(lobby.Players) != 1 {
		t.Errorf("Expected 1 player after leave, got %d", len(lobby.Players))
	}
}

func TestLobbyManager_Events(t *testing.T) {
	var joinCalled, leaveCalled, fullCalled, emptyCalled bool
	events := &LobbyEvents{
		OnPlayerJoin:  func(lobby *Lobby, player *Player) { joinCalled = true },
		OnPlayerLeave: func(lobby *Lobby, player *Player) { leaveCalled = true },
		OnLobbyFull:   func(lobby *Lobby) { fullCalled = true },
		OnLobbyEmpty:  func(lobby *Lobby) { emptyCalled = true },
	}
	manager := NewLobbyManagerWithEvents(events)

	lobby, err := manager.CreateLobby("EventLobby", 1, true, nil)
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}
	p := &Player{ID: "p1", Username: "Alice"}

	if err := manager.JoinLobby(lobby.ID, p); err != nil {
		t.Errorf("JoinLobby failed: %v", err)
	}
	if !joinCalled {
		t.Error("OnPlayerJoin event not called")
	}
	if !fullCalled {
		t.Error("OnLobbyFull event not called (should be full after 1 join)")
	}

	if err := manager.LeaveLobby(lobby.ID, p.ID); err != nil {
		t.Errorf("LeaveLobby failed: %v", err)
	}
	if !leaveCalled {
		t.Error("OnPlayerLeave event not called")
	}
	if !emptyCalled {
		t.Error("OnLobbyEmpty event not called")
	}
}
