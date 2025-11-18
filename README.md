# Decoupled Bidirectional gRPC Plugin Communication: The Missing Example

> **This is a demonstration project** that shows patterns for building production-ready plugin systems with HashiCorp's go-plugin. It fills gaps in existing documentation by showing how to build **decoupled, secure, and extensible** bidirectional communication between host and plugins.

## Why This Exists

Most go-plugin examples show simple unidirectional communication: host calls plugin. Done. But real-world plugin systems need more:

- **Plugins need to call back to the host** for controlled resource access
- **Security matters**: plugins shouldn't have direct filesystem/network access
- **Multiple plugins should share services** without code duplication
- **Adding new capabilities should be trivial**, not require refactoring

The examples in the go-plugin repo don't show these patterns clearly. This project does.

## What Makes This Different (and Cool)

### 1. **Clean Separation of Concerns**

We've separated infrastructure from business logic using a reusable `hostconn` package:

```go
// In your host - ONE LINE to setup host services for any plugin
hostconn.EstablishHostServices(plugin, hostServices, logger)

// That's it. No type casting, no broker management, no complexity.
```

**Compare this to typical implementations** where you manually:
- Type cast to broker-aware interfaces
- Register services with the broker
- Pass service IDs around
- Handle connection lifecycle

Our `hostconn` package handles all of that, making host service setup trivial.

### 2. **One Service Implementation, Multiple Plugins**

```go
// Create ONE host service implementation
hostServices := hostserve.NewHostServices(...)

// Share it with multiple plugins - one line each
hostconn.EstablishHostServices(plugin1, hostServices, logger)
hostconn.EstablishHostServices(plugin2, hostServices, logger)
```

The broker acts as a multiplexer, routing each plugin's calls to the same implementation through separate connections. This is powerful but not obvious from go-plugin docs.

### 3. **Optional Host Services**

Plugins that don't need host services? Just skip implementing the `HostConnection` interface:

```go
// Simple plugin - no host service boilerplate needed
type SimplePlugin struct {}

func (p *SimplePlugin) DoWork() string {
    return "done"
}
```

The helper functions gracefully handle both cases. No special conditionals needed.

### 4. **Easy Extensibility**

Adding a new host service function is trivial:

1. Add one RPC method to `hostserve.proto`:
   ```protobuf
   rpc GetEnv(GetEnvRequest) returns (GetEnvResponse);
   ```

2. Run `buf generate`

3. Implement the method:
   ```go
   func (h *HostServices) GetEnv(ctx context.Context, key string) string {
       return os.Getenv(key)
   }
   ```

That's it. All existing plugins can now call `GetEnv()`. No plugin code changes needed.

**Want to add a completely new service?** Same pattern - define proto, generate, implement. The `hostconn` infrastructure handles the connection plumbing.

### 5. **Client Identification for Capability-Based Security**

Here's where it gets really interesting for end users:

```go
// Plugin identifies itself in context
ctx = context.WithValue(ctx, "clientID", uuid.New())
entries, _ := hostServiceClient.ReadDir(ctx, "/sensitive-data")
```

```go
// Host service checks client capabilities
func (h *HostServices) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
    clientID := ctx.Value("clientID").(uuid.UUID)

    // Check what this client is allowed to access
    if !h.capabilities.CanAccess(clientID, path) {
        return nil, errors.New("access denied")
    }

    return os.ReadDir(path)
}
```

**Why This Matters for End Users:**

- **Sandboxing**: Each plugin gets only the permissions it declares
- **Audit trail**: Know exactly which plugin accessed what resource
- **Dynamic permissions**: Grant/revoke capabilities at runtime
- **Zero-trust plugins**: Plugins never touch the filesystem directly

**Real-world scenario**: A plugin marketplace where users install third-party plugins. Each plugin declares "I need to read config files" and the host enforces that it can ONLY read config files, not secrets or system files.

This is the foundation for building plugin systems users can trust.

## Recent Updates

**Latest improvements to the architecture:**

- **Streaming file operations**: Added infrastructure for streaming large files via `FileRead` (server streaming) and `FileWrite` (client streaming) to handle large files efficiently without loading everything into memory
- **File handle management**: Introduced `FileOpen`/`FileClose` endpoints with handle-based operations, enabling better resource management and explicit lifecycle control
- **Simplified architecture**: Removed `OpenRoots` complexity in favor of cleaner internal path confinement
- **Consistent naming**: Standardized method names (`FileOpen`, `FileClose`) for better API clarity
- **Dual access patterns**: Both simple unary operations (for small files) and handle-based streaming (for large files) are now supported

See the [Dual File Access Patterns](#dual-file-access-patterns) section for details on how to use both approaches.

## Quick Start

### Prerequisites
- Go 1.25+
- buf CLI (for protobuf generation)

### Build and Run

```bash
# Build everything
go build -o host .
go build -o plugins/filelister/filelister ./plugins/filelister
go build -o plugins/colorlister/colorlister ./plugins/colorlister

# Run the demo
./host
```

You'll see:
- The host spawning two plugins
- Plugins calling back to host services to read directories
- `filelister` demonstrating file handle operations (open, use, close)
- `colorlister` reading file contents with colored output and context propagation
- Clean shutdown with proper connection cleanup

## Architecture: The Big Picture

```
┌─────────────────────────────────────────────────────────────┐
│                        Host Process                          │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │         Shared Host Service Implementation          │    │
│  │         (ReadDir, ReadFile, WriteFile, GetEnv)      │    │
│  └──────────────────┬──────────────┬───────────────────┘    │
│                     │              │                         │
│       ┌─────────────┴──────┐  ┌───┴──────────────┐          │
│       │  Broker 1          │  │  Broker 2        │          │
│       │  (multiplexer)     │  │  (multiplexer)   │          │
│       └─────────┬──────────┘  └──────┬───────────┘          │
│                 │                    │                       │
└─────────────────┼────────────────────┼───────────────────────┘
                  │                    │
        ┌─────────┴──────────┐  ┌──────┴─────────┐
        │                    │  │                │
┌───────┼──────────┐  ┌──────┼──────────┐       │
│   Plugin 1       │  │   Plugin 2      │       │
│   (isolated)     │  │   (isolated)    │       │
│                  │  │                 │       │
│  Calls ReadDir() │  │ Calls ReadFile()│       │
└──────────────────┘  └─────────────────┘       │
        Both plugins securely call host services
        with their own capabilities/permissions
```

**Key insight**: Plugins are isolated processes. They can't access resources directly. They MUST go through host services, giving you complete control.

## What's Implemented

This demo includes:

**Two Example Plugins:**
- `filelister`: Lists files and demonstrates file handle operations (`FileOpen`/`FileClose`) for resource management
- `colorlister`: Reads files with colored output, demonstrates context propagation and client identification

**Host Services:**

*Simple Unary Operations:*
- `ReadDir(path)`: Read directory contents (unary RPC)
- `ReadFile(path)`: Read file contents (unary RPC)
- `WriteFile(path, data, perm)`: Write file (unary RPC)
- `GetEnv(key)`: Get environment variable

*File Handle Operations (for larger files and streaming):*
- `FileOpen(path, mode, perm)`: Open file and return handle
- `FileClose(handle)`: Close file handle
- `FileRead(handle, chunk_size)`: Read file in chunks (server streaming)
- `FileWrite(handle, chunks)`: Write file in chunks (client streaming)

**Infrastructure:**
- `hostconn` package: Reusable connection management for any plugin type
- Clean separation between business logic and infrastructure
- Proper connection lifecycle (setup → use → teardown)
- Thread-safe broker multiplexing
- File handle management for streaming operations

### Dual File Access Patterns

The project provides two ways to work with files:

**1. Simple Unary Operations** (for small files):
```go
// Read entire file at once
data, err := hostServiceClient.ReadFile(ctx, "/path/to/file")

// Write entire file at once
err = hostServiceClient.WriteFile(ctx, "/path/to/file", data, 0644)
```

**2. File Handle Operations** (for large files and streaming):
```go
// Open file and get handle
handle, size, err := hostServiceClient.FileOpen(ctx, "/path/to/file", READ_ONLY, 0)

// Read in chunks via streaming (when implemented)
stream, err := hostServiceClient.FileRead(ctx, handle, 8192)

// Close when done
err = hostServiceClient.FileClose(ctx, handle)
```

The file handle pattern enables:
- **Streaming large files** without loading everything into memory
- **Progressive processing** of file contents
- **Better resource management** with explicit open/close lifecycle
- **Future extensibility** for seeking, partial writes, etc.

**Note:** The streaming implementations for `FileRead` and `FileWrite` are defined in the proto and infrastructure is in place, but full implementations are still in progress.

## Project Structure

```
.
├── main.go                           # Host: spawns plugins, shares services
├── plugins/
│   ├── filelister/                   # Demo plugin 1
│   └── colorlister/                  # Demo plugin 2
├── shared/
│   ├── proto/                        # Service definitions
│   │   ├── filelister/v1/           # Plugin interface
│   │   └── hostserve/v1/            # Host services (add new services here)
│   ├── protogen/                     # Generated code (don't edit)
│   └── pkg/
│       ├── hostconn/                 # Reusable infrastructure (the magic)
│       ├── hostserve/                # Host service implementations
│       └── filelister/               # Plugin interface
├── buf.yaml                          # Proto module config
└── buf.gen.yaml                      # Code generation config
```

## How to Extend: Add a New Host Service Function

Adding new functionality is straightforward. Here's the general process:

**Step 1**: Edit `shared/proto/hostserve/v1/hostserve.proto` and add your new RPC:
```protobuf
service HostService {
  // ... existing methods ...
  rpc YourNewMethod(YourRequest) returns (YourResponse);  // NEW
}

message YourRequest {
  string param1 = 1;
  int32 param2 = 2;
}

message YourResponse {
  string result = 1;
  optional string error = 2;
}
```

**Step 2**: Regenerate code:
```bash
buf generate
```

**Step 3**: Implement the method:
- Add the business logic to `shared/pkg/hostserve/` (e.g., in `host_fs.go` for file operations)
- Add the gRPC server handler in `grpc_server_*.go`
- Add the gRPC client wrapper in `grpc_client_*.go`

**Done!** All plugins can now call your new method. No changes needed to:
- The broker setup
- The connection management
- Any plugin code (unless they want to use the new function)
- The `hostconn` infrastructure

## How to Add a New Plugin

**Minimal plugin (no host services needed):**

```go
type MyPlugin struct{}

func (p *MyPlugin) DoWork() string {
    return "I don't need host services"
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: handshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "my-plugin": &MyPluginGRPCWrapper{Impl: &MyPlugin{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

**Plugin with host services:**

```go
type MyPlugin struct {
    broker            *plugin.GRPCBroker
    hostServiceClient hostserve.IHostServices
    conn              *grpc.ClientConn
    connMutex         sync.Mutex
}

// Implement HostConnection interface
func (p *MyPlugin) SetBroker(broker *plugin.GRPCBroker) {
    p.broker = broker
}

func (p *MyPlugin) EstablishHostServices(hostServiceID uint32) {
    p.connMutex.Lock()
    defer p.connMutex.Unlock()

    conn, _ := p.broker.Dial(hostServiceID)
    p.conn = conn
    p.hostServiceClient = hostserve.NewHostServiceGRPCClient(
        hostservev1.NewHostServiceClient(conn))
}

func (p *MyPlugin) DisconnectHostServices() {
    p.connMutex.Lock()
    defer p.connMutex.Unlock()

    if p.conn != nil {
        p.conn.Close()
    }
}

// Now use host services in your business logic
func (p *MyPlugin) DoWork() (string, error) {
    entries, err := p.hostServiceClient.ReadDir(context.Background(), ".")
    return fmt.Sprintf("Found %d files", len(entries)), err
}
```

In your host:
```go
plugin := dispensePlugin("my-plugin")
hostconn.EstablishHostServices(plugin, hostServices, logger)  // One line!
```

## Security: Building Capability-Based Sandboxing

The client identification pattern demonstrated in `colorlister` is the foundation for real security:

### Current Implementation (Demo)
```go
// colorlister.go:31
ctx = context.WithValue(ctx, "client", "cl-plugin")
```

### Production Implementation (Conceptual)

**1. Plugin declares capabilities in manifest:**
```yaml
# plugins/myplugin/manifest.yaml
name: my-plugin
version: 1.0.0
capabilities:
  - read:config/**
  - write:output/**
  - env:API_KEY
```

**2. Host assigns UUID and loads capabilities:**
```go
clientID := uuid.New()
caps := loadCapabilities("plugins/myplugin/manifest.yaml")
capabilityManager.Register(clientID, caps)

ctx := context.WithValue(context.Background(), "clientID", clientID)
```

**3. Host services enforce capabilities:**
```go
func (h *HostServices) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
    clientID := ctx.Value("clientID").(uuid.UUID)

    if !capabilityManager.CanRead(clientID, path) {
        h.logger.Warn("Access denied", "client", clientID, "path", path)
        return nil, ErrAccessDenied
    }

    // Validate path is within allowed boundaries
    if !isPathAllowed(clientID, path) {
        return nil, ErrAccessDenied
    }

    return h.hostFS.ReadDir(ctx, path)
}
```

**4. Audit trail:**
```go
h.auditLog.Log(AuditEntry{
    ClientID:  clientID,
    Action:    "ReadDir",
    Resource:  path,
    Allowed:   true,
    Timestamp: time.Now(),
})
```

### What This Enables

**For End Users:**
- Install third-party plugins with confidence
- Know exactly what each plugin can access
- Revoke access without restarting
- Full audit trail of plugin behavior
- Plugins can't escalate privileges

**For Developers:**
- Clear security boundaries
- Easy to reason about access control
- Test plugins in isolation
- No risk of plugin compromising host

**Real-world example**: A text editor with plugin marketplace. A "file counter" plugin declares it needs `read:workspace/**`. It can't access your SSH keys, can't write files, can't access environment variables. If it tries, the host service denies the request and logs the attempt.

## Technical Deep Dives

### The Broker: How Multiplexing Works

Each plugin gets its own broker instance. When you call:
```go
broker.AcceptAndServe(serviceID, serverFunc)
```

The broker creates a gRPC server listening on a socket. When a plugin calls:
```go
conn, _ := broker.Dial(serviceID)
```

The broker connects to that socket. Multiple plugins can have service ID 1 because they're using different broker instances - each routes to the appropriate socket.

### Connection Ownership Model

**Critical distinction:**

- **Host owns servers**: Created via `broker.AcceptAndServe()`, managed by host
- **Plugin owns connections**: Created via `broker.Dial()`, cleaned up in `DisconnectHostServices()`

Plugins never have access to stop the host's servers. They only close their own connections.

### Why `hostconn` Package Matters

Without `hostconn`, your host code looks like:
```go
// Manual approach (verbose, error-prone)
grpcClient, ok := raw.(interface{ GetBroker() *plugin.GRPCBroker })
if !ok {
    // handle error
}
broker := grpcClient.GetBroker()
serviceID := broker.NextId()

serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
    s := grpc.NewServer(opts...)
    hostservev1.RegisterHostServiceServer(s, hostServices)
    return s
}
go broker.AcceptAndServe(serviceID, serverFunc)

if hc, ok := raw.(HostConnection); ok {
    hc.EstablishHostServices(serviceID)
}
```

With `hostconn`:
```go
hostconn.EstablishHostServices(raw, hostServices, logger)
```

This is what "reusable infrastructure" means. The complexity exists once, in a tested package, not repeated in every host implementation.

### Thread Safety Considerations

Host services can be called by multiple plugins concurrently. All implementations in this project use:
- `sync.Mutex` for connection management
- `os.OpenRoot()` which provides safe path confinement
- Context propagation for cancellation/timeouts

When extending, ensure your implementations are thread-safe.

## Common Patterns from This Codebase

**Pattern**: One service, multiple plugins (main.go:35-104)
```go
hostServices := hostserve.NewHostServices(...)
hostconn.EstablishHostServices(plugin1, hostServices, logger)
hostconn.EstablishHostServices(plugin2, hostServices, logger)
```

**Pattern**: Context-based client identification (colorlister.go:31)
```go
ctx = context.WithValue(ctx, "client", "cl-plugin")
result := hostServiceClient.ReadDir(ctx, dir)
```

**Pattern**: Safe file operations with internal root confinement (host_fs.go:48-62)
```go
// HostFS internally uses os.OpenRoot() for path confinement
func (hf *HostFS) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
    r, err := getRoot(path)
    if err != nil {
        return nil, err
    }
    defer closeRoot(r)
    return fs.ReadDir(r.FS(), ".")
}
```

**Pattern**: Connection lifecycle management
```go
// Setup - connect to host services
conn, _ := broker.Dial(serviceID)
plugin.conn = conn

// Use - call host services
entries := hostServiceClient.ReadDir(ctx, path)

// Teardown - clean up connection
if plugin.conn != nil {
    plugin.conn.Close()
}
```

**Pattern**: File handle lifecycle (for large file operations)
```go
// Open file and get handle
handle, size, _ := hostServiceClient.FileOpen(ctx, path, mode, perm)

// Use handle for operations
// ... read or write operations ...

// Always close handle when done
defer hostServiceClient.FileClose(ctx, handle)
```

## Comparison to go-plugin Examples

| Feature | go-plugin Examples | This Project |
|---------|-------------------|--------------|
| Bidirectional RPC | Shown but tightly coupled | Clean separation via `hostconn` |
| Multiple plugins | Not clearly demonstrated | Two plugins sharing services |
| Service registration | Manual broker management | One-line helper function |
| Connection lifecycle | Implicit or unclear | Explicit setup/teardown pattern |
| Security patterns | Not addressed | Client ID → capabilities foundation |
| Extensibility | Add service = lots of changes | Add service = edit proto, implement |
| Infrastructure reuse | Each example duplicates code | `hostconn` package works for all plugins |

## Why You Might Want This Pattern

**You're building a plugin system where:**
- Plugins come from different sources (marketplace, third-party, user scripts)
- Security matters (can't trust plugins with direct system access)
- You want to add capabilities over time without breaking plugins
- Multiple plugins should share services efficiently
- You need audit trails of plugin behavior

**You'll save time because:**
- The infrastructure is reusable across all plugin types
- Adding new host services is trivial (edit proto → implement)
- Connection management is handled consistently
- Security can be added incrementally via context values

## Protobuf Workflow

When modifying service definitions:

1. Edit `.proto` files in `shared/proto/`
2. Run `buf generate`
3. Implement new methods in corresponding Go files
4. **Never manually edit files in `shared/protogen/`**

The buf configuration ensures consistent code generation with proper Go module paths.

## Learning Resources

- [go-plugin Documentation](https://github.com/hashicorp/go-plugin) - Core framework
- [gRPC Go Basics](https://grpc.io/docs/languages/go/basics/) - Understanding gRPC
- [Protocol Buffers Guide](https://protobuf.dev/getting-started/gotutorial/) - Proto syntax

## FAQ

**Q: Is this production-ready?**
A: This is a demonstration/teaching project. The patterns are production-ready, but you'd want to add error handling, logging, metrics, testing, etc.

**Q: Can I use this code in my project?**
A: Yes, especially the `hostconn` package which is designed to be reusable. Treat this as a reference implementation.

**Q: Why not just give plugins direct filesystem access?**
A: Security and control. Plugins become sandboxed, auditable, and can't accidentally or maliciously access resources they shouldn't.

**Q: How does this handle plugin crashes?**
A: go-plugin provides process isolation. If a plugin crashes, the host continues running. You'd need to add recovery/restart logic.

**Q: Can plugins talk to each other?**
A: Not directly in this model. They'd need to go through host services. You could add a "message bus" host service for inter-plugin communication.

**Q: What about performance?**
A: gRPC is efficient. For high-frequency calls, you'd want connection pooling (examples in comments). The broker overhead is minimal.

## Contributing

This is an educational project. Improvements that clarify the patterns or demonstrate additional capabilities are welcome. Please maintain focus on teaching the architecture, not adding production features.

## License

This example is provided as-is for educational purposes.

---

**Questions?** This project exists to help the community understand production-ready patterns for go-plugin. If something is unclear, open an issue - your question helps improve the documentation.