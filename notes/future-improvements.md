# Future Improvements

## Transport/Adapter Layer Abstraction

Consider creating separate adapter libraries to handle networking/transport concerns for the core multiplayer-lobby package. For example:

- `multiplayer-lobby-ws`: WebSocket adapter
- `multiplayer-lobby-http`: HTTP/REST adapter
- `multiplayer-lobby-grpc`: gRPC adapter

**Benefits:**
- Keeps the core package transport-agnostic and focused on lobby logic
- Enables reusability and modularity—users can swap or add new transports easily
- Improves testability and separation of concerns

**Pattern:**
- Each adapter handles its own server setup and message parsing
- Adapters call into the core library for lobby operations
- Adapters translate core events/responses into the appropriate protocol messages

This approach is common in modern libraries and frameworks, and would make the multiplayer-lobby package more extensible and maintainable. 

## Persistent Session Storage

Consider supporting persistent session storage so that user sessions can survive server restarts. This could be implemented using a database or persistent key-value store, such as:

- **SQLite:** File-based SQL database, easy to use, no extra process, robust persistence. Recommended for most single-server setups.
- **Redis:** In-memory key-value store with optional disk persistence. Very fast, but requires running a Redis server.
- **PostgreSQL/MySQL:** Full-featured SQL databases, best if your app already uses them or needs advanced features.
- **BoltDB/BadgerDB:** Embedded, pure Go key-value stores, no external process, good for simple persistent storage.

**Recommendation:**
- Use SQLite for most single-server Go projects (simple, robust, portable).
- Use Redis if you want in-memory speed and optional persistence.
- Use Postgres/MySQL if you already have them in your stack.
- Use BoltDB for pure Go, embedded key-value needs.

This would allow users to reconnect and retain their session even after a server restart, improving reliability and user experience. 