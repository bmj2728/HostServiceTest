# Decoupled Bidirectional gRPC Plugin Communication: The Missing Example

> **This is a demonstration project** that shows patterns for building production-ready plugin systems with HashiCorp's go-plugin. It fills gaps in existing documentation by showing how to build **decoupled, secure, and extensible** bidirectional communication between host and plugins.

## Why This Exists

Most go-plugin examples show simple unidirectional communication: host calls plugin. Done. But real-world plugin systems need more (surprise!):

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

Adding a new host service function follows a clear pattern:

1. **Update proto and regenerate**:
   ```protobuf
   rpc GetEnv(GetEnvRequest) returns (GetEnvResponse);
   ```
   Then run `buf generate`

2. **Update the interface** in `shared/pkg/hostserve/host_service.go`:
   ```go
   type IHostServices interface {
       GetEnv(ctx context.Context, key string) (string, error)
       // ... other methods
   }
   ```

3. **Implement the business logic** in your host service:
   ```go
   func (h *HostServices) GetEnv(ctx context.Context, key string) (string, error) {
       return os.Getenv(key), nil
   }
   ```

4. **Add gRPC client wrapper** (matches host implementation signature):
   ```go
   func (c *HostServiceGRPCClient) GetEnv(ctx context.Context, key string) (string, error) {
       resp, err := c.client.GetEnv(ctx, &hostservev1.GetEnvRequest{Key: key})
       if err != nil { return "", err }
       return resp.Value, nil
   }
   ```

5. **Add gRPC server wrapper** (calls host implementation):
   ```go
   func (s *HostServiceGRPCServer) GetEnv(ctx context.Context, req *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {
       value, err := s.Impl.GetEnv(ctx, req.Key)
       return &hostservev1.GetEnvResponse{Value: value}, err
   }
   ```

That's it. All existing plugins can now call `GetEnv()`. No plugin code changes needed.

**Want to add a completely new service?** Same pattern - define proto, generate, implement. The `hostconn` infrastructure handles the connection plumbing.

### 5. **Client Identification for Capability-Based Security**

Here's where it gets really interesting for end users:

```go
// Plugin gets its client ID when establishing host services
func (p *Plugin) EstablishHostServices(hostServiceID uint32) (hostserve.ClientID, error) {
    conn, _ := p.broker.Dial(hostServiceID)
    client := hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))
    p.hostServiceClient = client
    return client.ClientID(), nil  
}

// Plugin just makes normal calls - client ID is automatically included
ctx := context.Background()
entries, _ := hostServiceClient.ReadDir(ctx, "/home/user", "/sensitive-data")
```

```go
// Host service automatically receives and can check client capabilities
func (h *HostServices) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {
    clientID := getClientIDFromContext(ctx)  // Extracted from gRPC metadata

    // Check what this client is allowed to access
    if !h.capabilities.CanAccess(clientID, rootDir, path) {
        return nil, errors.New("access denied")
    }

    return h.hostFS.ReadDir(ctx, rootDir, path)
}
```

**Why This Matters for End Users:**

- **Sandboxing**: Each plugin gets only the permissions it declares
- **Audit trail**: Know exactly which plugin accessed what resource
- **Dynamic permissions**: Grant/revoke capabilities at runtime
- **Zero-trust plugins**: Plugins never touch the filesystem directly

**Real-world scenario**: A plugin marketplace where users install third-party plugins. Each plugin declares "I need to read config files" and the host enforces that it can ONLY read config files, not secrets or system files. No sneaky business allowed.

This is the foundation for building plugin systems users can trust.

## Recent Updates

**Latest improvements to the architecture:**

- **File and directory management**: Added `Rename`, `Remove`, and `RemoveAll` operations for comprehensive file system management. These destructive operations follow the same security model and path confinement as other operations - use responsibly!
- **User and group identification**: Added `Getuid`, `Getgid`, `Geteuid`, `Getegid`, and `GetGroups` methods to retrieve user and group identifiers from the host system, enabling plugins to make security-aware decisions without direct system access
- **Consistent rootDir + path API pattern**: Refactored all file operations (`ReadDir`, `ReadFile`, `WriteFile`, `FileOpen`, `FileCreate`, `Stat`) to consistently accept a `rootDir` (must be absolute) and a `path` (absolute or relative, must not escape rootDir) instead of attempting to identify this data manually. This provides clearer security boundaries and better path confinement. See the [API Design Patterns](#api-design-patterns) section for details.
- **Convenience stat method**: Added `Stat` method as a simpler alternative to FileOpen -> FileStat -> FileClose pattern for retrieving file information when you just need metadata
- **Truly temporary files with automated cleanup**: Added `MkdirTemp` and `FileCreateTemp` methods that create temporary directories and files with server-side lifecycle tracking for guaranteed cleanup when connections close - unlike stdlib temps that persist until manual deletion
- **File positioning**: Implemented `FileSeek` method for standard file seeking operations (similar to `os.File.Seek`)
- **File information retrieval**: Added `FileStat` method to retrieve file metadata (size, mode, modification time) for open file handles - returns an object nearly identical to `fs.FileInfo`, but always returns nil for `Sys()`
- **Temp directory access**: Added `TempDir` method to retrieve the system temporary directory path
- **Additional filesystem operations**: Added `Mkdir`, `MkdirAll`, and `FileCreate` methods to provide complete directory creation and file initialization capabilities
- **Streaming file operations**: Fully implemented streaming for large files via `FileReader` (server streaming) and `FileWriter` (client streaming) to handle large files efficiently without loading everything into memory
- **Chunk size constraints**: `FileReader` enforces chunk sizes between 8KB (minimum for performance) and ~3.81MB (maximum due to gRPC 4MB message limits)
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
- `ReadDir(rootDir, path)`: Read directory contents (unary RPC)
- `ReadFile(rootDir, path)`: Read file contents (unary RPC)
- `WriteFile(rootDir, path, data, perm)`: Write file (unary RPC)
- `Stat(rootDir, path)`: Get file information (size, mode, modification time) for a path - convenience method that doesn't require file handle operations
- `Rename(rootDir, oldPath, newPath)`: Rename or move a file or directory (unary RPC)
- `Remove(rootDir, path)`: Delete a file or empty directory (unary RPC)
- `RemoveAll(rootDir, path)`: Recursively delete a directory and its contents (unary RPC) - **use with caution!**
- `Mkdir(rootDir, name, perm)`: Create a single directory (unary RPC)
- `MkdirAll(rootDir, path, perm)`: Create directory and all necessary parents (unary RPC)
- `MkdirTemp(rootDir, pattern)`: Create temporary directory with server-side tracking for automatic cleanup (unary RPC)
- `FileCreate(rootDir, path)`: Create or truncate a file (unary RPC)
- `GetEnv(key)`: Get environment variable
- `TempDir()`: Get system temporary directory path
- `UserCacheDir()`: Get user-specific cache directory (mirrors `os.UserCacheDir`)
- `UserConfigDir()`: Get user-specific configuration directory (mirrors `os.UserConfigDir`)
- `UserHomeDir()`: Get user's home directory (mirrors `os.UserHomeDir`)
- `Getuid()`: Get the user ID of the current process
- `Getgid()`: Get the group ID of the current process
- `Geteuid()`: Get the effective user ID of the current process
- `Getegid()`: Get the effective group ID of the current process
- `GetGroups()`: Get the list of supplementary group IDs for the current process

*File Handle Operations (for larger files and streaming):*
- `FileOpen(rootDir, path, mode, perm)`: Open file and return handle
- `FileCreateTemp(rootDir, pattern)`: Create temporary file with server-side tracking for automatic cleanup, returns handle
- `FileSeek(handle, offset, whence)`: Seek to position in file (similar to `os.File.Seek`)
- `FileStat(handle)`: Get file information (size, mode, modification time) for an open file handle - returns an `fs.FileInfo`-compatible object that always returns nil for `Sys()`
- `FileClose(handle)`: Close file handle
- `FileReader(handle, chunk_size)`: Read file in chunks via server streaming (fully implemented)
  - Requires `chunk_size` between 8KB and ~3.81MB
  - Returns an `io.Reader` for progressive file reading
- `FileWriter(handle, chunks)`: Write file in chunks via client streaming (fully implemented)

**Infrastructure:**
- `hostconn` package: Reusable connection management for any plugin type
- Clean separation between business logic and infrastructure
- Proper connection lifecycle (setup → use → teardown)
- Thread-safe broker multiplexing
- File handle management for streaming operations

### Temporary Files and Directories with Automated Cleanup

Unlike standard library temporary files that persist until manually deleted, this project implements **truly temporary** files and directories with server-side lifecycle management. When a plugin connection closes, all temporary resources created by that plugin are automatically cleaned up.

**Creating Temporary Directories:**

```go
// Create a temporary directory with a pattern
// The pattern works like os.MkdirTemp - "*" is replaced with random string
tempDir, err := hostServiceClient.MkdirTemp(ctx, "", "myapp-*")
if err != nil {
    log.Fatal(err)
}

// Use the temporary directory
log.Printf("Created temp dir: %s", tempDir)

// No cleanup needed! When the plugin disconnects, the directory and its
// contents are automatically removed by the host
```

**Creating Temporary Files:**

```go
// Create a temporary file and get a handle
// Pattern works like os.CreateTemp - "*" is replaced with random string
handle, err := hostServiceClient.FileCreateTemp(ctx, "", "data-*.json")
if err != nil {
    log.Fatal(err)
}

// Write to the temporary file
writer, err := hostServiceClient.FileWriter(ctx, handle)
if err != nil {
    log.Fatal(err)
}

data := []byte(`{"status": "processing"}`)
_, err = writer.Write(data)
writer.Close()

// Close the file handle
err = hostServiceClient.FileClose(ctx, handle)

// Automatic cleanup when plugin disconnects - no manual deletion needed!
```

**Getting the System Temp Directory:**

```go
// Get the system temporary directory path
tempDir, err := hostServiceClient.TempDir(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("System temp directory: %s", tempDir)
```

**Accessing User-Specific Directories:**

```go
// Get user cache directory (e.g., ~/.cache on Linux, ~/Library/Caches on macOS)
cacheDir, err := hostServiceClient.UserCacheDir(ctx)
if err != nil {
    log.Fatal(err)
}
log.Printf("User cache directory: %s", cacheDir)

// Get user config directory (e.g., ~/.config on Linux, ~/Library/Application Support on macOS)
configDir, err := hostServiceClient.UserConfigDir(ctx)
if err != nil {
    log.Fatal(err)
}
log.Printf("User config directory: %s", configDir)

// Get user home directory
homeDir, err := hostServiceClient.UserHomeDir(ctx)
if err != nil {
    log.Fatal(err)
}
log.Printf("User home directory: %s", homeDir)
```

These methods mirror Go's stdlib functions (`os.UserCacheDir`, `os.UserConfigDir`, `os.UserHomeDir`) and return platform-specific paths following OS conventions.

**Why This Matters:**

- **No resource leaks**: Plugins that crash or disconnect abnormally won't leave temporary files behind
- **Simplified plugin code**: No need to track and clean up temporary resources manually
- **Secure isolation**: Each plugin's temporary files are tracked separately and cleaned up independently
- **Better testing**: Test runs won't pollute the temp directory with leftover files

This is particularly valuable for long-running plugin systems where plugins may be loaded, unloaded, and reloaded multiple times throughout the host's lifetime.

### Dual File Access Patterns

The project provides two ways to work with files:

**1. Simple Unary Operations** (for small files):
```go
// Read entire file at once
// rootDir must be an absolute path, path can be absolute or relative (but must not escape rootDir)
data, err := hostServiceClient.ReadFile(ctx, "/home/user", "documents/file.txt")

// Write entire file at once
err = hostServiceClient.WriteFile(ctx, "/home/user", "documents/file.txt", data, 0644)

// Get file information without opening a handle (convenience method)
fileInfo, err := hostServiceClient.Stat(ctx, "/home/user", "documents/file.txt")
if err != nil {
    log.Fatal(err)
}
log.Printf("File size: %d, Mode: %v, ModTime: %v", fileInfo.Size(), fileInfo.Mode(), fileInfo.ModTime())

// Create a single directory
err = hostServiceClient.Mkdir(ctx, "/home/user", "newdir", 0755)

// Create directory and all necessary parents
err = hostServiceClient.MkdirAll(ctx, "/home/user", "path/to/nested/dirs", 0755)

// Create or truncate a file
err = hostServiceClient.FileCreate(ctx, "/home/user", "documents/newfile.txt")
```

**2. File Handle Operations** (for large files and streaming):
```go
// Open file and get handle
// rootDir must be an absolute path, path can be absolute or relative (but must not escape rootDir)
handle, size, err := hostServiceClient.FileOpen(ctx, "/home/user", "documents/largefile.bin", READ_ONLY, 0)

// Read in chunks via streaming
// chunk_size must be between 8KB (8192) and ~3.81MB (4000000)
chunkSize := uint32(64 * 1024) // 64KB is a good default
stream, err := hostServiceClient.FileReader(ctx, handle, chunkSize)

// Stream implements io.Reader, so you can use it directly
buffer := make([]byte, 1024)
for {
    n, err := stream.Read(buffer)
    if n > 0 {
        // Process the data in buffer[:n]
    }
    if err == io.EOF {
        break
    }
    if err != nil {
        // Handle error
        break
    }
}

// Seek to a specific position in the file
// whence: 0 = from start, 1 = from current position, 2 = from end
newOffset, err := hostServiceClient.FileSeek(ctx, handle, 1024, 0)
if err != nil {
    log.Fatal(err)
}
log.Printf("New file position: %d", newOffset)

// Get file information (metadata) for the open file
// Returns an fs.FileInfo-compatible object (Name, Size, Mode, ModTime, IsDir)
// Note: Sys() always returns nil for remote file handles
// TIP: If you only need file info and don't need to read/write, use Stat(path) instead
fileInfo, err := hostServiceClient.FileStat(ctx, handle)
if err != nil {
    log.Fatal(err)
}
log.Printf("File size: %d, Mode: %v, ModTime: %v", fileInfo.Size(), fileInfo.Mode(), fileInfo.ModTime())

// Close when done
err = hostServiceClient.FileClose(ctx, handle)
```

**Writing files with streaming (FileWriter):**

```go
// Open file for writing
// rootDir must be an absolute path, path can be absolute or relative (but must not escape rootDir)
handle, _, err := hostServiceClient.FileOpen(ctx, "/home/user", "output/data.bin", WRITE_ONLY, 0644)

// Write data in chunks via client streaming
data := []byte("large amount of data to write...")
chunkSize := 64 * 1024 // 64KB chunks - the Goldilocks size (not too small, not too large)

writer, err := hostServiceClient.FileWriter(ctx, handle)

// Write chunks
for i := 0; i < len(data); i += chunkSize {
    end := i + chunkSize
    if end > len(data) {
        end = len(data)
    }

    chunk := data[i:end]
    n, err := writer.Write(chunk)
    if err != nil {
        // Handle error
        break
    }
}

// Finish writing (flush any buffered data)
err = writer.Close()

// Close file handle
err = hostServiceClient.FileClose(ctx, handle)
```

The file handle pattern enables:
- **Streaming large files** without loading everything into memory
- **Progressive processing** of file contents
- **Better resource management** with explicit open/close lifecycle
- **Configurable chunk sizes** for optimal performance

**Chunk Size Guidelines:**

The `FileReader` operation requires a chunk size parameter with the following constraints:

- **Minimum: 8KB (8,192 bytes)** - Enforced to avoid excessive overhead and align with common buffer/block sizes
- **Maximum: ~3.81MB (4,000,000 bytes)** - gRPC enforces a 4MB message limit; this leaves room for message overhead
- **Recommended: 64KB-1MB** - Balances memory usage with network efficiency for most use cases

```go
// Common chunk size examples:
const (
    ChunkSize8KB   = 8 * 1024      // 8KB  - minimum allowed
    ChunkSize64KB  = 64 * 1024     // 64KB - good default for most files
    ChunkSize256KB = 256 * 1024    // 256KB - better for larger files
    ChunkSize1MB   = 1024 * 1024   // 1MB - maximum efficiency for very large files
    ChunkSize4MB   = 4000000       // ~3.81MB - maximum allowed
)
```

**Note:** See `plugins/filelister/filelister.go` for a complete working example of file handle operations with `FileReader` streaming.

### File Management Operations: With Great Power...

The host service API now includes operations that can modify and delete files and directories. While these are essential for building useful plugins, they require careful consideration:

**Destructive Operations:**
- `Rename(rootDir, oldPath, newPath)`: Renames or moves files and directories
- `Remove(rootDir, path)`: Deletes a single file or empty directory
- `RemoveAll(rootDir, path)`: Recursively deletes directories and all their contents

> **With great power there must also come great responsibility.** These operations can permanently delete data. Always validate paths and consider implementing confirmation mechanisms in production systems.

**Usage Examples:**

```go
// Rename a file safely
err := hostServiceClient.Rename(ctx, "/home/user", "old-name.txt", "new-name.txt")
if err != nil {
    log.Printf("Rename failed: %v", err)
}

// Remove a single file
err = hostServiceClient.Remove(ctx, "/home/user", "temp-file.txt")
if err != nil {
    log.Printf("Remove failed: %v", err)
}

// Remove an entire directory tree - USE WITH EXTREME CAUTION
// This will delete the directory and ALL its contents recursively
err = hostServiceClient.RemoveAll(ctx, "/home/user", "old-project-dir")
if err != nil {
    log.Printf("RemoveAll failed: %v", err)
}
```

**Security Considerations:**

All destructive operations respect the same `rootDir` + `path` security model:
- Paths cannot escape the specified `rootDir`
- The capability system can restrict which plugins can perform destructive operations
- All operations are logged for audit trails

**Best Practices:**

1. **Validate before destroying**: Always check what you're about to delete
2. **Use Remove for single files**: Prefer `Remove` over `RemoveAll` when you only need to delete one file
3. **Implement confirmations**: In interactive plugins, ask for user confirmation before destructive operations
4. **Leverage capabilities**: Configure your capability system to restrict destructive operations to trusted plugins only
5. **Check return values**: Always check for errors - they might indicate the path doesn't exist or lacks permissions

### User and Group Identification

For plugins that need to make security-aware decisions or understand the execution context, the host service provides access to user and group identifiers:

```go
// Get user and group IDs
uid, err := hostServiceClient.Getuid(ctx)      // Real user ID
gid, err := hostServiceClient.Getgid(ctx)      // Real group ID
euid, err := hostServiceClient.Geteuid(ctx)    // Effective user ID
egid, err := hostServiceClient.Getegid(ctx)    // Effective group ID
groups, err := hostServiceClient.GetGroups(ctx) // Supplementary group IDs

log.Printf("Running as UID %d (effective: %d), GID %d (effective: %d)", uid, euid, gid, egid)
log.Printf("Supplementary groups: %v", groups)
```

These methods mirror the standard Unix system calls and are useful for:
- Determining file ownership and permissions
- Making security decisions based on the executing user
- Logging and audit trails with user context
- Cross-platform user identification (returns appropriate values on Windows)

**Note:** These return the IDs of the **host process**, not the plugin process, since plugins are isolated and shouldn't need direct system access.

### API Design Patterns

The host service API follows two consistent patterns for resource access:

#### 1. RootDir + Path Pattern (Unary Operations & Resource Establishment)

All unary operations and methods that establish file handles use a **rootDir + path** pattern:

```go
// Pattern: operation(ctx, rootDir, path, ...other parameters)
data, err := hostServiceClient.ReadFile(ctx, rootDir, path)
err = hostServiceClient.WriteFile(ctx, rootDir, path, data, perm)
handle, size, err := hostServiceClient.FileOpen(ctx, rootDir, path, flags, perm)
```

**Requirements:**
- **rootDir**: MUST be an absolute path (e.g., `/home/user`, `/var/data`)
- **path**: CAN be absolute or relative, but MUST NOT escape the rootDir

**Why this pattern?**
- **Security**: Confines operations within a specific root directory, preventing path traversal attacks
- **Clarity**: Explicitly separates the security boundary (rootDir) from the target location (path)
- **Flexibility**: Plugins can pass absolute or relative paths while the host enforces confinement

**Examples:**
```go
// Both of these work - path can be relative or absolute
hostServiceClient.ReadFile(ctx, "/home/user", "documents/file.txt")
hostServiceClient.ReadFile(ctx, "/home/user", "/home/user/documents/file.txt")

// This will fail - rootDir must be absolute
hostServiceClient.ReadFile(ctx, "documents", "file.txt") // ERROR: rootDir not absolute
```

#### 2. Handle Pattern (Operations on Established Resources)

Once a file handle is established, all subsequent operations use only the **handle**:

```go
// First, establish the resource and get a handle
handle, size, err := hostServiceClient.FileOpen(ctx, rootDir, path, flags, perm)

// Then, operate using only the handle - no paths needed
reader, err := hostServiceClient.FileReader(ctx, handle, chunkSize)
offset, err := hostServiceClient.FileSeek(ctx, handle, offset, whence)
info, err := hostServiceClient.FileStat(ctx, handle)
err = hostServiceClient.FileClose(ctx, handle)
```

**Why this pattern?**
- **Performance**: Path resolution and security checks happen once at open time, not on every operation
- **Stateful operations**: Supports file position (seek), progressive reading/writing, and other stateful behavior
- **Resource management**: Explicit lifecycle with open/close makes resource tracking clear
- **Familiar**: Mirrors standard file I/O patterns (open → use → close)

**Key Differences:**

| Pattern | Use Case | Parameters | Security Check |
|---------|----------|------------|----------------|
| **RootDir + Path** | Unary ops, establishing resources | `(ctx, rootDir, path, ...)` | Every call |
| **Handle** | Operating on established resources | `(ctx, handle, ...)` | At establishment time |

This design provides both convenience (unary operations for simple cases) and efficiency (handle-based operations for complex workflows).

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

Adding new functionality is straightforward. Here's the complete process:

**Step 1: Update proto and regenerate**

Edit `shared/proto/hostserve/v1/hostserve.proto` and add your new RPC:
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

Then regenerate:
```bash
buf generate
```

**Step 2: Update the interface**

Add your method to `shared/pkg/hostserve/host_service.go`:
```go
type IHostServices interface {
    YourNewMethod(ctx context.Context, param1 string, param2 int32) (string, error)
    // ... other methods
}
```

**Step 3: Implement the business logic**

Add the implementation to `shared/pkg/hostserve/` (e.g., `host_fs.go` for file operations, or create a new file):
```go
func (h *HostServices) YourNewMethod(ctx context.Context, param1 string, param2 int32) (string, error) {
    // Your business logic here
    result := fmt.Sprintf("Processed %s with %d", param1, param2)
    return result, nil
}
```

**Step 4: Add gRPC client wrapper**

In `shared/pkg/hostserve/grpc_client_*.go`, add the client method that matches your interface signature:
```go
func (c *HostServiceGRPCClient) YourNewMethod(ctx context.Context, param1 string, param2 int32) (string, error) {
    resp, err := c.client.YourNewMethod(ctx, &hostservev1.YourNewMethodRequest{
        Param1: param1,
        Param2: param2,
    })
    if err != nil {
        return "", err
    }
    return resp.Result, nil
}
```

**Step 5: Add gRPC server wrapper**

In `shared/pkg/hostserve/grpc_server_*.go`, add the server method that calls your implementation:
```go
func (s *HostServiceGRPCServer) YourNewMethod(ctx context.Context, req *hostservev1.YourNewMethodRequest) (*hostservev1.YourNewMethodResponse, error) {
    result, err := s.Impl.YourNewMethod(ctx, req.Param1, req.Param2)
    if err != nil {
        return nil, err
    }
    return &hostservev1.YourNewMethodResponse{Result: result}, nil
}
```

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
    // rootDir must be an absolute path, path can be absolute or relative (but must not escape rootDir)
    entries, err := p.hostServiceClient.ReadDir(context.Background(), "/home/user", ".")
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
func (h *HostServices) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {
    clientID := ctx.Value("clientID").(uuid.UUID)

    if !capabilityManager.CanRead(clientID, rootDir, path) {
        h.logger.Warn("Access denied", "client", clientID, "rootDir", rootDir, "path", path)
        return nil, ErrAccessDenied
    }

    // Validate rootDir is within allowed boundaries
    if !isRootDirAllowed(clientID, rootDir) {
        return nil, ErrAccessDenied
    }

    return h.hostFS.ReadDir(ctx, rootDir, path)
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

### The Broker: How Multiplexing Works (The Secret Sauce)

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

**Pattern**: Safe file operations with internal root confinement (host_fs.go:81-121)
```go
// HostFS uses rootDir + path pattern with os.OpenInRoot() for path confinement
func (hf *HostFS) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {
    // Validate rootDir is absolute
    if !filepath.IsAbs(rootDir) {
        return nil, fmt.Errorf("rootDir is not absolute: %s", rootDir)
    }

    // Convert path to relative if needed
    rel, err := absToRel(rootDir, path)
    if err != nil {
        return nil, err
    }

    // Open in root for safe path confinement
    f, err := os.OpenInRoot(rootDir, rel)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    return f.ReadDir(0)
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

The buf configuration ensures consistent code generation with proper Go module paths. (And yes, you really should never edit the generated files. We know it's tempting. Don't.)

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
A: Security and control. Plugins become sandboxed, auditable, and can't accidentally or maliciously access resources they shouldn't. Think of it as a velvet rope for your filesystem - plugins can look, but only touch what they're allowed to.

**Q: How does this handle plugin crashes?**
A: go-plugin provides process isolation. If a plugin crashes, the host continues running. You'd need to add recovery/restart logic.

**Q: Can plugins talk to each other?**
A: Not directly in this model - they're like kids at separate lunch tables. They'd need to go through host services. You could add a "message bus" host service for inter-plugin communication if you want them to pass notes.

**Q: What about performance?**
A: gRPC is efficient - basically, it's fast. For high-frequency calls, you'd want connection pooling (examples in comments). The broker overhead is minimal. You're more likely to be bottlenecked by your business logic than the infrastructure.

## Contributing

This is an educational project. Improvements that clarify the patterns or demonstrate additional capabilities are welcome. Please maintain focus on teaching the architecture, not adding production features.

## License

This example is provided as-is for educational purposes.

---

**Questions?** This project exists to help the community understand production-ready patterns for go-plugin. If something is unclear, open an issue - your question helps improve the documentation.