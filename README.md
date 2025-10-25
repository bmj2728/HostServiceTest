# Bidirectional gRPC Plugin Example

This project demonstrates **bidirectional gRPC communication** using [HashiCorp's go-plugin](https://github.com/hashicorp/go-plugin) framework. Unlike simple plugin architectures where the host only calls into plugins, this example shows how plugins can call back to the host for services while maintaining process isolation.

## Table of Contents

- [Why This Example?](#why-this-example)
- [Architecture Overview](#architecture-overview)
- [Key Concepts](#key-concepts)
- [Project Structure](#project-structure)
- [Building and Running](#building-and-running)
- [How It Works](#how-it-works)
- [Communication Flow](#communication-flow)
- [Extending This Example](#extending-this-example)
- [Common Patterns](#common-patterns)

## Why This Example?

Most go-plugin examples show simple unidirectional communication: host → plugin. However, many real-world use cases require plugins to call back to the host for:

- **Controlled resource access** (files, network, databases)
- **Shared services** (logging, configuration, authentication)
- **Security boundaries** (sandboxed plugins with limited privileges)

This example demonstrates the complete pattern using a file listing plugin that must call back to the host to read directories.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Host Process                          │
│                                                               │
│  ┌──────────────────┐         ┌─────────────────────────┐  │
│  │                  │         │   Host Services         │  │
│  │  Plugin Client   │         │  (HostService gRPC)     │  │
│  │                  │         │                         │  │
│  │  - Spawns plugin │         │  - ReadDir()            │  │
│  │  - Calls plugin  │         │  - Uses os.OpenRoot()   │  │
│  │    methods       │         │                         │  │
│  └──────────────────┘         └─────────────────────────┘  │
│           │                              ▲                   │
│           │                              │                   │
│           ├──────── Broker ──────────────┤                   │
│           │      (gRPC Broker)           │                   │
│           │   - Manages connections      │                   │
│           │   - Allocates service IDs    │                   │
│           ▼                              │                   │
└───────────┼──────────────────────────────┼───────────────────┘
            │                              │
            │ Unix Socket / Named Pipe     │
            │                              │
┌───────────┼──────────────────────────────┼───────────────────┐
│           ▼                              │                   │
│  ┌──────────────────┐         ┌─────────────────────────┐  │
│  │                  │         │                         │  │
│  │  Plugin Server   │         │  Plugin Implementation  │  │
│  │  (FileLister)    │         │                         │  │
│  │                  │         │  - ListFiles()          │  │
│  │  - Receives      │────────▶│  - Dials broker         │  │
│  │    host service  │         │  - Calls ReadDir()      │  │
│  │    connection ID │         │    on host service      │  │
│  └──────────────────┘         └─────────────────────────┘  │
│                                                               │
│                       Plugin Process                          │
└─────────────────────────────────────────────────────────────┘
```

## Key Concepts

### 1. The Broker

The **broker** is go-plugin's mechanism for bidirectional RPC. It allows both sides to:
- Register gRPC servers (via `AcceptAndServe`)
- Connect to servers on the other side (via `Dial`)

Each service is identified by a `uint32` ID that both sides must agree on.

### 2. Service IDs

Service IDs are allocated dynamically using an atomic counter:

```go
var brokerIDCounter uint32 = 0

func allocateServiceID() uint32 {
    return atomic.AddUint32(&brokerIDCounter, 1)
}
```

This ensures multiple services can coexist without ID conflicts.

### 3. Two Directions of Communication

**Host → Plugin (Plugin Services)**
- Defined in `shared/proto/filelister/v1/filelister.proto`
- Example: `FileLister.List()`
- Plugin implements the service, host calls it

**Plugin → Host (Host Services)**
- Defined in `shared/proto/hostserve/v1/hostserve.proto`
- Example: `HostService.ReadDir()`
- Host implements the service, plugin calls it

### 4. Connection Ownership Model

Understanding who owns what is critical:

**Host Side:**
- Creates gRPC **server** via `broker.AcceptAndServe(serviceID, serverFunc)`
- The server is owned and managed by the host
- Host is responsible for stopping its own servers on shutdown

**Plugin Side:**
- Creates gRPC **connections** via `broker.Dial(serviceID)`
- Each connection should be tracked for cleanup
- Plugin must close its connections in `DisconnectHostServices()`

**Important:** The plugin never has access to the host's server and shouldn't try to stop it!

## Project Structure

```
.
├── main.go                          # Host process entry point
├── plugins/
│   └── filelister/
│       ├── filelister.go           # Plugin implementation
│       └── manifest.yaml           # Plugin metadata
├── shared/
│   ├── proto/                      # Protocol Buffer definitions
│   │   ├── filelister/v1/
│   │   │   └── filelister.proto   # Plugin service definition
│   │   └── hostserve/v1/
│   │       └── hostserve.proto    # Host service definition
│   ├── protogen/                   # Generated Go code (DO NOT EDIT)
│   └── pkg/
│       ├── filelister/             # Plugin interface & gRPC wrappers
│       │   ├── filelister.go      # Interface and GRPCClient
│       │   └── filelister_grpc.go # GRPCServer implementation
│       └── hostserve/              # Host service implementation
│           ├── host_services.go   # Core implementation
│           └── host_services_grpc.go # gRPC wrapper
├── buf.yaml                        # Buf configuration for proto
├── buf.gen.yaml                    # Buf code generation config
├── CLAUDE.md                       # Development guide
└── README.md                       # This file
```

## Building and Running

### Prerequisites

- Go 1.25+
- buf CLI (for regenerating protobuf code)

### Build

```bash
# Build the host process
go build -o host .

# Build the plugin
go build -o plugins/filelister/filelister ./plugins/filelister
```

### Run

```bash
# Run the host (it will automatically spawn the plugin)
./host
```

**Expected Output:**
```
2025-10-25T17:22:23.408-0400 [DEBUG] host: starting plugin: path=./plugins/filelister/filelister args=["./plugins/filelister/filelister"]
2025-10-25T17:22:23.408-0400 [DEBUG] host: plugin started: path=./plugins/filelister/filelister pid=32126
2025-10-25T17:22:23.409-0400 [DEBUG] host: waiting for RPC address: plugin=./plugins/filelister/filelister
2025-10-25T17:22:23.412-0400 [DEBUG] host.filelister: plugin address: address=/tmp/plugin570982019 network=unix timestamp=2025-10-25T17:22:23.412-0400
2025-10-25T17:22:23.412-0400 [DEBUG] host: using plugin: version=1
2025-10-25T17:22:23.413-0400 [INFO]  host: Host service registered with broker: id=1
2025-10-25T17:22:23.413-0400 [DEBUG] host.filelister: 2025-10-25T17:22:23.413-0400 [INFO]  Established host services: id=1
.claude   true
.git   true
.idea   true
CLAUDE.md   false
README.md   false
buf.gen.yaml   false
buf.yaml   false
go.mod   false
go.sum   false
host   false
internal   true
main.go   false
plugins   true
shared   true
2025-10-25T17:22:23.415-0400 [INFO]  host: Successfully listed files
.claude-d
.git-d
.idea-d
CLAUDE.md-f
README.md-f
buf.gen.yaml-f
buf.yaml-f
go.mod-f
go.sum-f
host-f
internal-d
main.go-f
plugins-d
shared-d
2025-10-25T17:22:23.415-0400 [INFO]  host: Shutting down plugin
```

The output shows:
- DEBUG logs tracking plugin lifecycle (starting, connecting, etc.)
- Host service registration with broker ID
- Intermediate output showing directory entries with boolean flags (true=directory, false=file)
- Final formatted file listing with `-d` suffix for directories and `-f` suffix for regular files
- Clean shutdown sequence

### Regenerate Protobuf Code

When modifying `.proto` files:

```bash
buf generate
```

This regenerates all files in `shared/protogen/`. Never edit these files manually.

## How It Works

### Step-by-Step Flow

#### 1. Host Starts Plugin

```go
// main.go
client := plugin.NewClient(&plugin.ClientConfig{
    HandshakeConfig: handshakeConfig,
    Plugins: map[string]plugin.Plugin{
        "fl-plugin": &filelister.FileListerGRPCPlugin{},
    },
    Cmd: exec.Command("./plugins/filelister/filelister"),
    AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
})
```

#### 2. Host Registers Host Service with Broker

```go
// Host allocates a service ID and starts serving
hostServices := &hostserve.HostServices{}
hostServiceID, err := grpcClientImpl.SetupHostService(hostServices)
// Returns: hostServiceID = 1 (dynamically allocated)
```

Inside `SetupHostService`:
```go
func (c *GRPCClient) SetupHostService(hostServices hostserve.IHostServices) (uint32, error) {
    // Allocate unique ID
    serviceID := atomic.AddUint32(&brokerIDCounter, 1)

    // Register with broker
    go c.broker.AcceptAndServe(serviceID, func(opts []grpc.ServerOption) *grpc.Server {
        server := grpc.NewServer(opts...)
        hostservev1.RegisterHostServiceServer(server, &hostserve.HostServiceGRPCServer{
            Impl: hostServices,
        })
        return server
    })

    return serviceID, nil
}
```

#### 3. Host Tells Plugin About the Service ID

```go
fileLister.EstablishHostServices(hostServiceID)
```

This sends the ID to the plugin via gRPC:
```protobuf
rpc EstablishHostServices(HostServiceRequest) returns (Empty);

message HostServiceRequest {
  uint32 host_service = 1;  // The broker service ID
}
```

#### 4. Plugin Stores the Service ID

```go
// plugins/filelister/filelister.go
func (f *FileLister) EstablishHostServices(hostServiceID uint32) {
    f.hostServices = hostServiceID  // Store for later use
}
```

#### 5. Host Calls Plugin Method

```go
entries, err := fileLister.ListFiles(".", hostServiceID)
```

#### 6. Plugin Dials Back to Host Service

```go
// plugins/filelister/filelister.go
func (f *FileLister) ListFiles(dir string, hostService uint32) ([]string, error) {
    // Connect to host service using the broker
    conn, err := f.broker.Dial(hostService)
    if err != nil {
        return nil, err
    }

    // Track this connection for cleanup (important!)
    f.connMutex.Lock()
    f.connections = append(f.connections, conn)
    f.connMutex.Unlock()

    // Create client for host service
    hsClient := hostserve.NewHostServiceGRPCClient(
        hostservev1.NewHostServiceClient(conn))

    // Call back to host to read directory
    dirEntries, err := hsClient.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    // Process entries: add suffix to indicate file type
    var entries []string
    for _, entry := range dirEntries {
        if entry.IsDir() {
            entries = append(entries, entry.Name()+"-d")
        } else {
            entries = append(entries, entry.Name()+"-f")
        }
    }

    return entries, nil
}
```

**Note:**
- Connections are tracked rather than immediately closed so they can be properly cleaned up in `DisconnectHostServices()`
- The plugin adds `-d` suffix for directories and `-f` suffix for regular files to demonstrate that plugins can process and transform data received from host services

#### 7. Host Service Executes Request

```go
// shared/pkg/hostserve/host_services.go
func (hs *HostServices) ReadDir(path string) ([]fs.DirEntry, error) {
    // Secure filesystem access using os.OpenRoot
    r, err := os.OpenRoot(path)
    if err != nil {
        return nil, err
    }
    defer r.Close()

    return fs.ReadDir(r.FS(), ".")
}
```

#### 8. Results Flow Back Through the Stack

```
Host Service → Broker → Plugin → Broker → Host Client
```

#### 9. Cleanup on Shutdown

```go
// Host initiates shutdown
fileLister.DisconnectHostServices()
client.Kill()  // Terminates plugin process
```

```go
// Plugin cleans up its connections
func (f *FileLister) DisconnectHostServices() {
    f.connMutex.Lock()
    defer f.connMutex.Unlock()

    for _, conn := range f.connections {
        conn.Close()
    }
    f.connections = nil
}
```

The host manages its own server lifecycle separately from the plugin shutdown.

## Communication Flow

### Sequence Diagram

```
Host                Broker              Plugin
  |                   |                   |
  |--Start Plugin---->|                   |
  |                   |<---Register-------|
  |                   |                   |
  |--AcceptAndServe-->|                   |
  |   (ID=1)          |                   |
  |                   |                   |
  |--EstablishHostServices(1)------------>|
  |                   |                   |
  |                   |   Plugin stores ID=1
  |                   |                   |
  |--ListFiles(".")---------------------->|
  |                   |                   |
  |                   |<----Dial(1)-------|
  |                   |                   |
  |<---ReadDir("/")---|-------------------|
  |                   |                   |
  |---[entries]-------|------------------>|
  |                   |                   |
  |<------[entries]----------------------|
  |                   |                   |
```

## Extending This Example

### Adding a New Host Service

1. **Define the service in protobuf:**

```protobuf
// shared/proto/newservice/v1/newservice.proto
service NewService {
  rpc DoSomething(Request) returns (Response);
}
```

2. **Generate code:**
```bash
buf generate
```

3. **Implement the service:**

```go
// shared/pkg/newservice/newservice.go
type NewService struct {}

func (s *NewService) DoSomething(req Request) (Response, error) {
    // Implementation
}
```

4. **Register with broker in host:**

```go
newServiceID, err := grpcClientImpl.SetupNewService(newServiceImpl)
fileLister.EstablishNewService(newServiceID)
```

5. **Use in plugin:**

```go
conn, err := f.broker.Dial(f.newServiceID)
client := newservev1.NewNewServiceClient(conn)
resp, err := client.DoSomething(ctx, req)
```

### Adding a New Plugin

1. **Define plugin service in protobuf:**

```protobuf
// shared/proto/myplugin/v1/myplugin.proto
service MyPlugin {
  rpc DoWork(WorkRequest) returns (WorkResponse);
}
```

2. **Create plugin interface:**

```go
// shared/pkg/myplugin/myplugin.go
type MyPlugin interface {
    DoWork(input string) (string, error)
    SetBroker(broker *plugin.GRPCBroker)
}
```

3. **Implement gRPC wrappers** (similar to `FileListerGRPCPlugin`)

4. **Implement the plugin:**

```go
// plugins/myplugin/myplugin.go
type MyPluginImpl struct {
    broker *plugin.GRPCBroker
}

func (p *MyPluginImpl) DoWork(input string) (string, error) {
    // Can dial back to host services here
    return result, nil
}
```

5. **Register in plugin main:**

```go
pluginMap := map[string]plugin.Plugin{
    "my-plugin": &myplugin.MyPluginGRPCPlugin{Impl: impl},
}
plugin.Serve(&plugin.ServeConfig{
    HandshakeConfig: handshakeConfig,
    Plugins:         pluginMap,
    GRPCServer:      plugin.DefaultGRPCServer,
})
```

## Common Patterns

### Pattern 1: Multiple Host Services

```go
// Host setup
fileServiceID, _ := client.SetupFileService(fileService)
dbServiceID, _ := client.SetupDatabaseService(dbService)
cacheServiceID, _ := client.SetupCacheService(cacheService)

// Pass all IDs to plugin
plugin.EstablishServices(fileServiceID, dbServiceID, cacheServiceID)
```

### Pattern 2: Service Discovery

Instead of passing individual IDs, pass a service registry:

```protobuf
message ServiceRegistry {
  map<string, uint32> services = 1;  // name -> broker ID
}
```

```go
registry := map[string]uint32{
    "file":  fileServiceID,
    "db":    dbServiceID,
    "cache": cacheServiceID,
}
plugin.EstablishServices(registry)
```

### Pattern 3: Lazy Service Connection

Don't dial the broker until the service is actually needed:

```go
type LazyHostServiceClient struct {
    broker    *plugin.GRPCBroker
    serviceID uint32
    client    hostservev1.HostServiceClient
    once      sync.Once
}

func (l *LazyHostServiceClient) getClient() hostservev1.HostServiceClient {
    l.once.Do(func() {
        conn, _ := l.broker.Dial(l.serviceID)
        l.client = hostservev1.NewHostServiceClient(conn)
    })
    return l.client
}
```

### Pattern 4: Connection Pooling

For high-throughput scenarios, maintain a connection pool:

```go
type HostServicePool struct {
    broker    *plugin.GRPCBroker
    serviceID uint32
    pool      chan *grpc.ClientConn
}

func (p *HostServicePool) Get() (*grpc.ClientConn, error) {
    select {
    case conn := <-p.pool:
        return conn, nil
    default:
        return p.broker.Dial(p.serviceID)
    }
}

func (p *HostServicePool) Put(conn *grpc.ClientConn) {
    select {
    case p.pool <- conn:
    default:
        conn.Close()  // Pool full, close connection
    }
}
```

## Security Considerations

### Process Isolation

Plugins run in separate processes, providing:
- Memory isolation
- Crash isolation
- Resource limit enforcement via OS

### Controlled Access

The host service pattern allows fine-grained control:

```go
type RestrictedHostService struct {
    allowedPaths []string
}

func (s *RestrictedHostService) ReadDir(path string) ([]fs.DirEntry, error) {
    if !s.isPathAllowed(path) {
        return nil, errors.New("access denied")
    }
    // ... perform operation
}
```

### Sandboxing with os.OpenRoot

The example uses `os.OpenRoot()` which provides path confinement:

```go
// Plugin can only access within the opened root
r, err := os.OpenRoot("/allowed/directory")
// Plugin cannot escape this directory boundary
```

## Troubleshooting

### Common Issues

**1. "Failed to dial host service"**
- Ensure the host service ID is passed correctly to the plugin
- Check that `AcceptAndServe` is called before the plugin tries to dial
- Verify the broker is the same instance on both sides

**2. "Assignment count mismatch"**
- Check that `SetupHostService` returns `(uint32, error)` not just `error`
- Ensure you're handling both return values

**3. "Plugin process not starting"**
- Verify the plugin binary exists and is executable
- Check the path in `exec.Command()` is correct
- Review plugin logs for startup errors

**4. "Service not implemented"**
- Ensure your gRPC server embeds `UnimplementedXXXServer`
- Verify all required methods are implemented
- Check that `RegisterXXXServer` is called

**5. "DisconnectHostServices signature confusion"**
- **Wrong:** `DisconnectHostServices(server *grpc.Server)` - Plugin doesn't have access to host's server
- **Right:** `DisconnectHostServices()` - Plugin closes its own connections
- The plugin manages **connections** (via `Dial`), not the host's **server** (via `AcceptAndServe`)

### Debug Logging

Enable detailed logging:

```go
logger := hclog.New(&hclog.LoggerOptions{
    Name:   "host",
    Output: os.Stdout,
    Level:  hclog.Debug,  // Change to Debug
})
```

## Performance Considerations

### Broker Overhead

Each broker dial creates a new gRPC connection. For high-frequency calls:
- Use connection pooling (see patterns above)
- Consider batching requests
- Cache connections when possible

### Serialization Cost

Protobuf is efficient, but large messages still have cost:
- Stream large datasets instead of single messages
- Use compression for large payloads
- Consider pagination for list operations

## References

- [go-plugin Documentation](https://github.com/hashicorp/go-plugin)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [Buf CLI](https://buf.build/)

## License

This example is provided as-is for educational purposes.

## Contributing

Improvements and clarifications to this example are welcome! Please ensure any changes maintain the clarity of the bidirectional communication pattern.

---

**Questions or Issues?** This example was created to help the community understand bidirectional go-plugin patterns. If something is unclear, please open an issue with your questions.
