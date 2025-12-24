#!/usr/bin/env python3
"""
Python plugin entry point for go-plugin.

This demonstrates how a Python plugin can integrate with go-plugin's
gRPC-based plugin architecture. The plugin communicates via stdio just
like the Go plugins do.
"""

import sys
import os
import grpc
from concurrent import futures
import time

# Add protogen_py to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', '..', 'shared', 'protogen_py'))

from filelister.v1 import filelister_pb2_grpc

# Import plugin implementation
try:
    from plugin import FileListerPlugin
except ImportError:
    # When running as PyInstaller bundle
    from src.plugin import FileListerPlugin


def validate_handshake():
    """
    Validate go-plugin handshake.

    go-plugin uses magic cookie values to ensure the host and plugin
    are compatible. This is a simple security measure.
    """
    # These match the Go handshake config
    expected_key = "TEST_KEY"
    expected_value = "TEST_VALUE"

    # go-plugin sets these as environment variables
    actual_value = os.getenv(expected_key)

    if actual_value != expected_value:
        print(f"[Python Plugin] Handshake failed: expected {expected_key}={expected_value}, got {actual_value}", file=sys.stderr)
        sys.exit(1)

    print(f"[Python Plugin] Handshake validated successfully", file=sys.stderr)


def main():
    """
    Main entry point for the Python plugin.

    go-plugin communicates via stdio:
    - Plugin prints connection info to stdout
    - Host reads this to connect via gRPC
    - All logging goes to stderr
    """
    print("[Python Plugin] Starting...", file=sys.stderr)

    # Validate the handshake
    validate_handshake()

    # Create gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    # Register our FileLister service (no broker needed - simplified version)
    plugin_impl = FileListerPlugin()
    filelister_pb2_grpc.add_FileListerServicer_to_server(plugin_impl, server)

    # Note: We don't register GRPCBroker service since this plugin doesn't use it
    # The "Method not found!" error from the broker is harmless - it's just the host
    # probing to see if we support bidirectional communication

    # Listen on a Unix socket (like Go plugins do)
    import tempfile
    socket_path = tempfile.mktemp(prefix='plugin', suffix='')
    socket_addr = f'unix://{socket_path}'
    server.add_insecure_port(socket_addr)

    print(f"[Python Plugin] Starting gRPC server on {socket_addr}...", file=sys.stderr)
    server.start()

    # go-plugin protocol: print connection info to stdout
    # Format: <protocol_version>|<protocol_type>|<network>|<address>|<protocol>|
    # Must match: 1|1|unix|/tmp/socketpath|grpc|
    connection_info = f"1|1|unix|{socket_path}|grpc|"
    print(connection_info, flush=True)

    print(f"[Python Plugin] Server started, waiting for connections...", file=sys.stderr)
    print(f"[Python Plugin] Connection info: {connection_info}", file=sys.stderr)

    try:
        # Keep the server running until terminated
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("[Python Plugin] Shutting down...", file=sys.stderr)
        server.stop(0)


if __name__ == '__main__':
    main()
