package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	lobby "github.com/jonosm/multiplayer-lobby"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsConn wraps a websocket connection to implement lobby.Conn
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsConn) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

// connManager tracks active connections
type connManager struct {
	mu    sync.RWMutex
	conns map[string]*wsConn
}

func newConnManager() *connManager {
	return &connManager{
		conns: make(map[string]*wsConn),
	}
}

func (cm *connManager) Add(userID string, conn *wsConn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[userID] = conn
}

func (cm *connManager) Remove(userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conns, userID)
}

func (cm *connManager) Get(userID string) (*wsConn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, ok := cm.conns[userID]
	return conn, ok
}

func main() {
	sessionManager := lobby.NewSessionManager()
	
	connMgr := newConnManager()

	events := &lobby.LobbyEvents{
		Broadcaster: func(userID string, message interface{}) {
			if conn, ok := connMgr.Get(userID); ok {
				conn.WriteJSON(message)
			}
		},
		OnPlayerJoin: func(l *lobby.Lobby, p *lobby.Player) {
			log.Printf("Player %s joined lobby %s", p.Username, l.Name)
		},
		OnPlayerLeave: func(l *lobby.Lobby, p *lobby.Player) {
			log.Printf("Player %s left lobby %s", p.Username, l.Name)
		},
		OnLobbyDeleted: func(l *lobby.Lobby) {
			log.Printf("Lobby %s deleted", l.Name)
		},
	}

	lobbyManager := lobby.NewLobbyManagerWithEvents(events)

	deps := &lobby.HandlerDeps{
		SessionManager: sessionManager,
		LobbyManager:   lobbyManager,
		ConnToUserID:   make(map[interface{}]string),
	}

	router := lobby.NewMessageRouter()
	router.SetupDefaultHandlers(deps)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}
		defer conn.Close()

		ws := &wsConn{conn: conn}
		var userID string

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				break
			}

			var msg struct {
				Action string `json:"action"`
			}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			if err := router.Dispatch(ws, message); err != nil {
				log.Printf("Dispatch error: %v", err)
			}

			if newUserID, ok := deps.ConnToUserID[ws]; ok && userID == "" {
				userID = newUserID
				connMgr.Add(userID, ws)
			}
		}

		if userID != "" {
			connMgr.Remove(userID)
			sessionManager.RemoveSession(userID)
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Connect via WebSocket at ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
