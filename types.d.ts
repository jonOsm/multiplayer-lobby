// TypeScript definitions for the multiplayer lobby library

export interface Player {
  user_id: string;
  username: string;
  ready: boolean;
  can_start_game?: boolean;
}

export interface Lobby {
  id: string;
  name: string;
  max_players: number;
  players: Player[];
  state: 'waiting' | 'in_game' | 'finished';
  metadata?: Record<string, any>;
  public: boolean;
}

// Request payload types
export interface RegisterUserPayload {
  username: string;
}

export interface CreateLobbyPayload {
  name: string;
  max_players: number;
  public: boolean;
  user_id: string;
  metadata?: Record<string, any>;
}

export interface JoinLobbyPayload {
  lobby_id: string;
  user_id: string;
}

export interface LeaveLobbyPayload {
  lobby_id: string;
  user_id: string;
}

export interface SetReadyPayload {
  lobby_id: string;
  user_id: string;
  ready: boolean;
}

export interface ListLobbiesPayload {
  // Empty payload
}

export interface StartGamePayload {
  lobby_id: string;
  user_id: string;
}

export interface GetLobbyInfoPayload {
  lobby_id: string;
}

export interface LogoutPayload {
  user_id: string;
}

// WebSocket message types
export interface RegisterUserRequest {
  action: 'register_user';
  data: RegisterUserPayload;
}

export interface CreateLobbyRequest {
  action: 'create_lobby';
  data: CreateLobbyPayload;
}

export interface JoinLobbyRequest {
  action: 'join_lobby';
  data: JoinLobbyPayload;
}

export interface LeaveLobbyRequest {
  action: 'leave_lobby';
  data: LeaveLobbyPayload;
}

export interface SetReadyRequest {
  action: 'set_ready';
  data: SetReadyPayload;
}

export interface ListLobbiesRequest {
  action: 'list_lobbies';
  data: ListLobbiesPayload;
}

export interface StartGameRequest {
  action: 'start_game';
  data: StartGamePayload;
}

export interface GetLobbyInfoRequest {
  action: 'get_lobby_info';
  data: GetLobbyInfoPayload;
}

export interface LogoutRequest {
  action: 'logout';
  data: LogoutPayload;
}

// Response types
export interface RegisterUserResponse {
  action: 'user_registered';
  user_id: string;
  username: string;
}

export interface LobbyStateResponse {
  action: 'lobby_state';
  lobby_id: string;
  players: Player[];
  state: string;
  metadata?: Record<string, any>;
}

export interface LobbyListResponse {
  action: 'lobby_list';
  lobbies: string[];
}

export interface LobbyInfoResponse {
  action: 'lobby_info';
  lobby_id: string;
  name: string;
  players: Player[];
  state: string;
  max_players: number;
  public: boolean;
}

export interface ErrorResponse {
  action: 'error';
  code: string;
  message: string;
  details?: string;
}

// Session event types
export interface SessionCreatedEvent {
  event: 'session_created';
  user_id: string;
  username: string;
}

export interface SessionReconnectedEvent {
  event: 'session_reconnected';
  user_id: string;
  username: string;
}

export interface SessionRemovedEvent {
  event: 'session_removed';
  user_id: string;
  username: string;
}

// Union types
export type WebSocketMessage = 
  | RegisterUserRequest
  | CreateLobbyRequest
  | JoinLobbyRequest
  | LeaveLobbyRequest
  | SetReadyRequest
  | ListLobbiesRequest
  | StartGameRequest
  | GetLobbyInfoRequest
  | LogoutRequest;

export type WebSocketResponse = 
  | RegisterUserResponse
  | LobbyStateResponse
  | LobbyListResponse
  | LobbyInfoResponse
  | ErrorResponse
  | SessionCreatedEvent
  | SessionReconnectedEvent
  | SessionRemovedEvent;

// Error codes
export type ErrorCode = 
  | 'USER_NOT_FOUND'
  | 'USER_INACTIVE'
  | 'USERNAME_TAKEN'
  | 'INVALID_USERNAME'
  | 'LOBBY_NOT_FOUND'
  | 'LOBBY_FULL'
  | 'LOBBY_NOT_WAITING'
  | 'PLAYER_NOT_IN_LOBBY'
  | 'PLAYER_ALREADY_IN_LOBBY'
  | 'LOBBY_ALREADY_EXISTS'
  | 'NOT_ENOUGH_PLAYERS'
  | 'NOT_ALL_PLAYERS_READY'
  | 'CANNOT_START_GAME'
  | 'INVALID_MESSAGE'
  | 'UNKNOWN_ACTION'
  | 'INVALID_REQUEST'
  | 'INTERNAL_ERROR'
  | 'SERVICE_UNAVAILABLE';

// Connection interface for transport-agnostic design
export interface Conn {
  WriteJSON(v: any): Promise<void> | void;
}

// Message handler interface
export interface MessageHandler {
  (conn: Conn, msg: IncomingMessage): Promise<void> | void;
}

// Incoming message interface
export interface IncomingMessage {
  action: string;
  data: any;
}

// Middleware interface
export interface Middleware {
  (next: MessageHandler): MessageHandler;
}

// Message router interface
export interface MessageRouter {
  Handle(action: string, handler: MessageHandler): void;
  Use(middleware: Middleware): void;
  Dispatch(conn: Conn, rawMsg: string | Buffer): Promise<void> | void;
} 