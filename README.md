# Decoupled Bidirectional gRPC Plugin Communication

![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)
![Python Version](https://img.shields.io/badge/Python-3.9%2B-3776AB?logo=python&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-Protocol-244c5a?logo=grpc)
![Protobuf](https://img.shields.io/badge/Protobuf-3-4285F4?logo=google&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue)
![Polyglot](https://img.shields.io/badge/Polyglot-Go%20%2B%20Python-blueviolet)
![Architecture](https://img.shields.io/badge/Architecture-Plugin%20System-orange)
![Status](https://img.shields.io/badge/Status-Demo%20%2F%20Reference-yellow)

> **A demonstration project** showing production-ready patterns for building plugin systems with HashiCorp's go-plugin. Fills gaps in existing documentation by demonstrating **decoupled, secure, and extensible** bidirectional communication between host and plugins.

```
┌─────────────────────────────────────────────────────────────────────┐
│  🚀 Production-Ready Patterns  |  🔒 Sandboxed Execution            │
│  🌐 Cross-Language Support     |  📦 Reusable Infrastructure        │
│  🎯 One Service, Many Plugins  |  ⚡ Concurrent & Thread-Safe       │
│  🖥️  Interactive TUI Interface  |  🧪 Comprehensive Test Suite       │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.25+
- buf CLI (for protobuf generation)
- Python 3.9+ (optional, for building Python plugin)

### Build and Run

```bash
# Build everything
go build -o hst .
go build -o plugins/filelister/filelister ./plugins/filelister
go build -o plugins/colorlister/colorlister ./plugins/colorlister
go build -o plugins/hostdemo/hostdemo ./plugins/hostdemo

# Optional: Build Python plugin
cd plugins/pylelister && ./build.sh && cd ../..

# Run with interactive TUI
./hst

# Or run CLI demo
./hst --no-tui
```

The **TUI (Text User Interface)** provides an interactive menu for exploring plugins, running functions with custom inputs, and viewing scrollable output. Perfect for understanding how the system works!

## What You'll See

### Interactive TUI Mode

```
┌─────────────────────────────────────┐
│ Host Service Plugin Demo System     │
├─────────────────────────────────────┤
│ > File Lister (Go)                  │
│   Color Lister (Go)                 │
│   Python Lister                     │
│   Host Demo                         │
│   Run Concurrency Demo              │
│   Reset State                       │
│   Shutdown                          │
└─────────────────────────────────────┘
```

Navigate with arrow keys or vim-style (j/k), select plugins, provide inputs, and see results in a scrollable view.

### What This Demonstrates

✅ **4 example plugins** (3 Go, 1 Python) showing real-world patterns
✅ **60+ host service operations** covering filesystem, processes, and more
✅ **Automatic client identification** for capability-based security
✅ **Cross-language compatibility** with Python plugins
✅ **Concurrent plugin execution** demonstrating thread-safe broker multiplexing
✅ **Interactive TUI** for exploring capabilities
✅ **Production patterns** that scale from demos to real systems

## Why This Project Exists

Most go-plugin examples show simple unidirectional communication: host calls plugin. Done. But real-world plugin systems need more:

- **Plugins calling back to the host** for controlled resource access
- **Security**: plugins shouldn't have direct filesystem/network access
- **Multiple plugins sharing services** without code duplication
- **Easy extensibility** without refactoring

The examples in the go-plugin repo don't show these patterns clearly. This project does.

## Documentation

📚 **[Visit the Wiki](../../wiki)** for complete documentation:

- **[Home](../../wiki/Home)** - Project overview and quick start
- **[Getting Started](../../wiki/Getting-Started)** - Detailed setup and build instructions
- **[TUI Guide](../../wiki/TUI-Guide)** - Interactive interface navigation
- **[Plugin Development](../../wiki/Plugin-Development)** - Create your own plugins
- **[Cross-Language Plugins](../../wiki/Cross-Language-Plugins)** - Python plugins and polyglot architecture
- **[API Reference](../../wiki/API-Reference)** - Complete host services API (60+ operations)
- **[Architecture](../../wiki/Architecture)** - Technical deep dives, broker mechanics, security patterns

## Key Features

### 1. Clean Separation of Concerns

```go
// In your host - ONE LINE to setup host services for any plugin
hostconn.EstablishHostServices(plugin, hostServices, logger)
```

The `hostconn` package handles all broker complexity, making host service setup trivial.

### 2. One Service Implementation, Multiple Plugins

```go
// Create ONE host service implementation
hostServices := hostserve.NewHostServices(...)

// Share it with multiple plugins - one line each
hostconn.EstablishHostServices(plugin1, hostServices, logger)
hostconn.EstablishHostServices(plugin2, hostServices, logger)
```

The broker multiplexes requests from different plugins to the same implementation.

### 3. Optional Host Services

Plugins that don't need host services? Just skip implementing the `HostConnection` interface. No boilerplate needed.

### 4. Easy Extensibility

Adding a new host service function:
1. Update proto and regenerate (`buf generate`)
2. Update the interface
3. Implement business logic
4. Add gRPC wrappers

All existing plugins can now call your new method. No plugin code changes needed.

### 5. Automatic Client Identification

Every plugin operation is logged with client ID, enabling audit trails and capability-based security:

```
[INFO]  host: ReadDir request from client:
  client=019b05f0-a32c-7b3d-a0f1-5087d232174f
  clientOwner=filelister
  request=019b05f0-a343-7753-93aa-117a437c5321
[INFO]  host: ReadDir completed successfully:
  duration_us=120 success=true
```

## Example Plugins

**Go Plugins (with host services):**
- **`hostdemo`** - Demonstration suite (GetEnv, system info, temp file lifecycle)
- **`filelister`** - File handle operations, streaming, resource management
- **`colorlister`** - Path-based operations, colored output, temp files

**Python Plugin (cross-language demo):**
- **`pylelister`** - Demonstrates protocol compatibility across languages (~21MB self-contained executable)

## Architecture

```
┌──────────────────────────────────────────┐
│           Host Process                    │
│  ┌────────────────────────────────┐      │
│  │  Shared Host Service           │      │
│  │  (File I/O, Env, Process...)   │      │
│  └────┬──────────┬──────────┬─────┘      │
│       │          │          │            │
│  ┌────┴───┐ ┌───┴────┐ ┌───┴────┐       │
│  │Broker 1│ │Broker 2│ │Broker 3│       │
│  └────┬───┘ └───┬────┘ └───┬────┘       │
└───────┼─────────┼──────────┼────────────┘
        │         │          │
  ┌─────┴───┐ ┌──┴────┐ ┌───┴─────┐
  │Go Plugin│ │Go Plug│ │Go Plugin│
  │  #1     │ │  #2   │ │  #3     │
  └─────────┘ └───────┘ └─────────┘
```

- **One service, many plugins** - Shared implementation, broker multiplexing
- **Process isolation** - Plugins run as separate processes
- **Capability-based security** - Client identification enables access control

## Project Structure

```
.
├── hst                              # Host executable
├── main.go                          # Host implementation with TUI
├── plugins/                         # Example plugins
│   ├── filelister/                  # Go plugin - file operations
│   ├── colorlister/                 # Go plugin - colored output
│   ├── hostdemo/                    # Go plugin - demo suites
│   └── pylelister/                  # Python plugin - cross-language
├── shared/
│   ├── proto/                       # Protobuf definitions
│   ├── protogen/                    # Generated Go code
│   ├── protogen_py/                 # Generated Python code
│   └── pkg/
│       ├── hostconn/                # Reusable infrastructure
│       ├── hostserve/               # Host service implementations
│       ├── filelister/              # FileLister interface
│       └── hostdemo/                # HostDemo interface
└── wiki/                            # Documentation
```

## Testing

Comprehensive test suite for host services:

```bash
# Run all tests
go test ./shared/pkg/hostserve

# With coverage
go test ./shared/pkg/hostserve -cover

# With race detector
go test ./shared/pkg/hostserve -race
```

Current coverage: 23.7% of statements

Tests include:
- Client tracking with concurrency (100 goroutines)
- File handle management with concurrency (50 goroutines)
- Temp path tracking and cleanup
- Environment and process operations
- File system operations
- Helper functions and conversions

See `shared/pkg/hostserve/TEST_README.md` for details.

## Recent Updates

**TUI Implementation (2025-12-27):**
- Interactive text-based interface for exploring plugins
- Menu-driven navigation with vim-style keys
- Input collection for plugin functions
- Scrollable output views
- Concurrency demo showing interleaved execution

**Streamlined Demo Architecture (2025-12-27):**
- Unified plugin lifecycle management
- New HostDemo service with easy-to-call suite methods
- Separation of demonstration logic from implementation patterns

**Cross-Language Support:**
- Python plugin demonstrating protocol compatibility
- Self-contained PyInstaller build (no Python installation needed)
- Foundation for future Python SDK

**Enhanced Logging and Metrics:**
- Structured request/completion logging
- Client ID and request ID tracking
- Microsecond-precision duration measurements
- Comprehensive audit trails

## Contributing

This is an educational project. We welcome contributions that:
- Clarify existing patterns
- Demonstrate additional capabilities
- Improve test coverage
- Contribute to Python SDK (`goplugin-py`)

Please maintain focus on teaching the architecture, not adding production features.

## License

MIT License - provided as-is for educational purposes.

## Learning Resources

- **[Project Wiki](../../wiki)** - Complete documentation
- [go-plugin Documentation](https://github.com/hashicorp/go-plugin) - Core framework
- [gRPC Go Basics](https://grpc.io/docs/languages/go/basics/) - Understanding gRPC
- [Protocol Buffers Guide](https://protobuf.dev/) - Proto syntax

---

**Questions?** This project exists to help the community understand production-ready patterns for go-plugin. Visit the [wiki](../../wiki) or open an issue if anything is unclear.
