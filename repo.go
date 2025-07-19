package lobby

import "sync"

// LobbyRepository defines the interface for lobby storage backends.
type LobbyRepository interface {
	CreateLobby(lobby *Lobby) error
	GetLobby(id LobbyID) (*Lobby, bool)
	ListLobbies() []*Lobby
	UpdateLobby(lobby *Lobby) error
	DeleteLobby(id LobbyID) error
}

// InMemoryLobbyRepo is a thread-safe in-memory implementation of LobbyRepository.
type InMemoryLobbyRepo struct {
	mu      sync.Mutex
	lobbies map[LobbyID]*Lobby
}

// NewInMemoryLobbyRepo creates a new in-memory lobby repository.
func NewInMemoryLobbyRepo() *InMemoryLobbyRepo {
	return &InMemoryLobbyRepo{
		lobbies: make(map[LobbyID]*Lobby),
	}
}

// CreateLobby stores a new lobby. Returns ErrLobbyExists if the lobby already exists.
func (r *InMemoryLobbyRepo) CreateLobby(lobby *Lobby) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.lobbies[lobby.ID]; exists {
		return ErrLobbyExists
	}
	r.lobbies[lobby.ID] = lobby
	return nil
}

// GetLobby retrieves a lobby by ID.
func (r *InMemoryLobbyRepo) GetLobby(id LobbyID) (*Lobby, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	lobby, exists := r.lobbies[id]
	return lobby, exists
}

// ListLobbies returns all lobbies in the repository.
func (r *InMemoryLobbyRepo) ListLobbies() []*Lobby {
	r.mu.Lock()
	defer r.mu.Unlock()
	lobbies := make([]*Lobby, 0, len(r.lobbies))
	for _, l := range r.lobbies {
		lobbies = append(lobbies, l)
	}
	return lobbies
}

// UpdateLobby updates an existing lobby. Returns ErrLobbyNotFound if the lobby does not exist.
func (r *InMemoryLobbyRepo) UpdateLobby(lobby *Lobby) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.lobbies[lobby.ID]; !exists {
		return ErrLobbyNotFound
	}
	r.lobbies[lobby.ID] = lobby
	return nil
}

// DeleteLobby removes a lobby by ID. Returns ErrLobbyNotFound if the lobby does not exist.
func (r *InMemoryLobbyRepo) DeleteLobby(id LobbyID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.lobbies[id]; !exists {
		return ErrLobbyNotFound
	}
	delete(r.lobbies, id)
	return nil
}

// Error variables for common repo errors.
var (
	ErrLobbyExists   = &RepoError{"lobby already exists"}
	ErrLobbyNotFound = &RepoError{"lobby not found"}
)

// RepoError represents a repository error.
type RepoError struct {
	msg string
}

// Error returns the error message.
func (e *RepoError) Error() string { return e.msg }
