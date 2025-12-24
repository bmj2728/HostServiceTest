# Python Plugin for go-plugin Demo

This directory contains a **Python implementation** of the FileLister plugin, demonstrating cross-language compatibility in the go-plugin architecture.

## What This Demonstrates

- **Cross-language plugins**: Python plugin works alongside Go plugins seamlessly
- **Host services are optional**: This plugin operates independently, showcasing that not all plugins need host callbacks
- **Self-contained deployment**: Bundles Python interpreter + dependencies into single executable
- **Zero host dependencies**: No Python installation required on host system
- **gRPC protocol compatibility**: Same protobuf definitions work across languages

## Important: Simplified Version

**This plugin does NOT use host services.** Instead, it accesses the filesystem directly, demonstrating that:

1. **Host services are optional** - plugins can be self-contained
2. **Cross-language gRPC works** - the core go-plugin protocol is language-agnostic
3. **Bidirectional broker complexity** - requires additional Python SDK development (future work)

### Why Not Host Services?

The go-plugin broker protocol for bidirectional communication (plugins calling back to host) requires:
- Implementing the `GRPCBroker.StartStream` protocol in Python
- Handling connection multiplexing
- Managing the knock/ack handshake

This would require building a **Python SDK for go-plugin** - a significant undertaking beyond the scope of this demo. For now, this plugin demonstrates that:
- Python plugins can integrate with go-plugin
- The stdio handshake protocol works
- gRPC service definitions are portable
- Multiple languages can coexist in the same plugin system

## Future Work: Full Python SDK

To support host services from Python plugins, we'd need to build `goplugin-py`:
```python
# Future API (not yet implemented)
import goplugin

class MyPlugin(FileListerServicer):
    def __init__(self, broker):
        self.broker = broker  # Handles StartStream protocol

    def EstablishHostServices(self, request, context):
        # Dial host service via broker
        conn = self.broker.Dial(request.host_service)
        self.host_stub = HostServiceStub(conn)

goplugin.serve({
    'filelister': MyPlugin
})
```

This would handle the broker protocol, connection management, and proper cleanup - mirroring what go-plugin provides for Go.

## Architecture (Current Simplified Version)

```
┌─────────────────────────────────────┐
│      Go Host Process                │
│  - Spawns plugins as subprocesses   │
│  - Provides HostService (unused)    │
└───────┬────────────────────┬────────┘
        │                    │
   ┌────┴──────┐      ┌──────┴──────────┐
   │ Go Plugin │      │ Python Plugin   │
   │ (binary)  │      │ (PyInstaller)   │
   │           │      │                 │
   │ Uses      │      │ Independent     │
   │ HostService│     │ file access     │
   └───────────┘      └─────────────────┘
                       │ Bundles:        │
                       │ - Python 3.13   │
                       │ - grpcio        │
                       │ - protobuf      │
                       │ - plugin code   │
                       └─────────────────┘
```

## Building

### Prerequisites

- Python 3.9+ (for building only - not needed to run!)
- pip
- buf (for generating protobuf stubs)

### Build Steps

```bash
# 1. Generate Python protobuf stubs (if not already done)
cd ../..
buf generate

# 2. Build the plugin executable
cd plugins/pylelister
./build.sh
```

This creates `dist/pylelister` - a self-contained executable binary (~21MB).

## What the Build Does

The `build.sh` script:
1. Creates an isolated Python virtual environment
2. Installs dependencies (grpcio, protobuf, pyinstaller)
3. Uses PyInstaller to bundle everything into a single executable
4. Result: `dist/pylelister`

The executable includes:
- Python 3.13 interpreter (embedded)
- Plugin code (`src/main.py`, `src/plugin.py`)
- Generated protobuf stubs from `shared/protogen_py/`
- All Python dependencies

## Running

The Python plugin is executed **exactly like a Go plugin** from the host's perspective:

```go
// In Go host code
exec.Command("./plugins/pylelister/dist/pylelister")
```

The host doesn't need to know it's a Python plugin - it just sees a gRPC-speaking subprocess.

## Implementation Details

### Handshake

Uses hardcoded handshake matching Go plugins:
```python
expected_key = "TEST_KEY"
expected_value = "TEST_VALUE"
```

### stdio Protocol

go-plugin uses stdio for initial connection:
1. Plugin starts, validates handshake via environment variables
2. Plugin starts gRPC server on Unix socket
3. Plugin prints connection info to stdout: `1|1|unix|/tmp/socketpath|grpc|`
4. Host reads this and connects via gRPC

### File Operations

Python plugin accesses filesystem directly:

```python
def List(self, request, context):
    full_path = os.path.join(request.root_dir, request.path)
    entries = os.listdir(full_path)
    # Format and return entries
```

**Note:** This means the plugin has direct filesystem access, unlike Go plugins that use host services. This is a tradeoff of the simplified approach.

## File Structure

```
plugins/pylelister/
├── src/
│   ├── __init__.py           # Package marker
│   ├── main.py               # Entry point, stdio protocol
│   └── plugin.py             # FileLister implementation (no host services)
├── dist/
│   └── pylelister            # Final executable (created by build)
├── requirements.txt          # Python dependencies
├── pylelister.spec           # PyInstaller configuration
├── build.sh                  # Build script
└── README.md                 # This file
```

## Dependencies

- **grpcio**: gRPC runtime for Python
- **grpcio-tools**: Protocol buffer compiler
- **protobuf**: Protocol buffer runtime
- **pyinstaller**: Bundles Python + deps into executable

These are only needed for **building**, not for running the final executable.

## Comparison to Go Plugins

| Aspect | Go Plugin | Python Plugin |
|--------|-----------|---------------|
| Language | Go | Python |
| Build tool | `go build` | PyInstaller |
| Binary size | ~5-8MB | ~21MB |
| Startup time | Fast | Slightly slower |
| Runtime | Native | Embedded interpreter |
| Host integration | Identical | Identical |
| Host services | ✅ Yes | ❌ No (requires SDK) |
| Direct FS access | ❌ Via host | ✅ Direct |

From the **host's connection perspective**, they're identical. From the **functionality perspective**, Python version is limited until a proper SDK exists.

## Testing

```bash
# Build the plugin
./build.sh

# The host will automatically use it if configured in main.go
cd ../..
./host
```

The plugin will:
- Start up and validate handshake
- Connect via Unix socket
- List files using direct filesystem access
- Return results with `[Python Plugin]` prefix

## Production Considerations

**Self-contained Benefits:**
- No "Python version hell" - plugin controls its Python version
- No dependency conflicts with other plugins
- Easier distribution - single file
- Works on systems without Python

**Current Limitations:**
- No host service callbacks (direct filesystem access instead)
- Larger binary size (~21MB vs ~5-8MB for Go)
- Slightly slower startup (embedded interpreter initialization)

**Security Note:** Since this plugin accesses the filesystem directly rather than through host services, it's not sandboxed like the Go plugins. A full implementation with host services would restore the security model.

## Why This Still Matters

Even though this plugin doesn't use host services, it demonstrates:

1. **Protocol compatibility**: go-plugin's stdio handshake and gRPC connection work with Python
2. **Polyglot architecture**: Multiple languages can coexist in the same system
3. **Optional complexity**: Not all plugins need bidirectional communication
4. **Foundation for future work**: This is a stepping stone toward full Python SDK

For plugins that don't need host callbacks (e.g., formatters, analyzers, simple transforms), this approach works perfectly.

## Next Steps for Full Integration

To build a complete Python SDK for go-plugin:

1. **Implement broker client** - Handle `GRPCBroker.StartStream` protocol
2. **Connection management** - Track dialable services and multiplexing
3. **Helper functions** - Simplify plugin authoring (like Go's `plugin.Serve`)
4. **Graceful shutdown** - Handle cleanup signals properly
5. **Testing** - Comprehensive tests against go-plugin host
6. **Documentation** - Usage examples and API docs

This would enable Python plugins to use host services just like Go plugins do, unlocking the full security and architectural benefits of the decoupled host service pattern.
