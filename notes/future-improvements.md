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