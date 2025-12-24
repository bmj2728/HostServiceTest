"""
FileLister plugin implementation - simplified version without host services.

This demonstrates that:
1. Python plugins can work with go-plugin
2. Host services are optional (not all plugins need them)
3. Cross-language gRPC communication works seamlessly

This plugin lists files by accessing the filesystem directly, showcasing
a plugin that doesn't require host service callbacks.
"""

import sys
import os

# Add protogen_py to path so we can import generated protobuf code
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', '..', 'shared', 'protogen_py'))

from filelister.v1 import filelister_pb2, filelister_pb2_grpc


class FileListerPlugin(filelister_pb2_grpc.FileListerServicer):
    """
    Simple FileLister plugin implementation in Python.

    This plugin demonstrates cross-language compatibility without the complexity
    of bidirectional host service connections. It accesses the filesystem directly,
    showing that plugins can be self-contained.
    """

    def __init__(self):
        """Initialize the plugin."""
        print("[Python Plugin] Initialized (no host services)", file=sys.stderr)

    def EstablishHostServices(self, request, context):
        """
        Host service establishment - not implemented.

        This plugin doesn't use host services, demonstrating that they're optional.
        We return an empty response to indicate "no host services needed".
        """
        print("[Python Plugin] EstablishHostServices called - not needed for this plugin", file=sys.stderr)

        # Return empty response - this is valid and indicates the plugin
        # doesn't require host services
        return filelister_pb2.HostServiceResponse(
            client_id="",  # Empty client_id = no host services
            error=None
        )

    def List(self, request, context):
        """
        List files in a directory by accessing filesystem directly.

        This demonstrates a plugin that operates independently without
        needing to call back to the host.
        """
        try:
            root_dir = request.root_dir
            path = request.path

            print(f"[Python Plugin] List called: root_dir={root_dir}, path={path}", file=sys.stderr)

            # Join paths safely
            if os.path.isabs(path):
                full_path = path
            else:
                full_path = os.path.join(root_dir, path)

            # Normalize the path
            full_path = os.path.normpath(full_path)

            print(f"[Python Plugin] Reading directory: {full_path}", file=sys.stderr)

            # Check if path exists
            if not os.path.exists(full_path):
                error_msg = f"Path does not exist: {full_path}"
                print(f"[Python Plugin] Error: {error_msg}", file=sys.stderr)
                return filelister_pb2.FileListResponse(error=error_msg)

            # Check if it's a directory
            if not os.path.isdir(full_path):
                error_msg = f"Path is not a directory: {full_path}"
                print(f"[Python Plugin] Error: {error_msg}", file=sys.stderr)
                return filelister_pb2.FileListResponse(error=error_msg)

            # List directory contents
            entries = []
            try:
                for entry_name in os.listdir(full_path):
                    entry_path = os.path.join(full_path, entry_name)

                    # Add directory indicator
                    if os.path.isdir(entry_path):
                        entries.append(f"[DIR] {entry_name}")
                    else:
                        entries.append(entry_name)

            except PermissionError as e:
                error_msg = f"Permission denied: {e}"
                print(f"[Python Plugin] Error: {error_msg}", file=sys.stderr)
                return filelister_pb2.FileListResponse(error=error_msg)

            # Add a marker to show this came from Python
            entries.insert(0, "=== [Python Plugin] Directory Listing ===")

            print(f"[Python Plugin] Found {len(entries) - 1} entries", file=sys.stderr)

            return filelister_pb2.FileListResponse(entry=entries)

        except Exception as e:
            error_msg = f"Unexpected error: {e}"
            print(f"[Python Plugin] Error in List: {error_msg}", file=sys.stderr)
            import traceback
            traceback.print_exc(file=sys.stderr)
            return filelister_pb2.FileListResponse(error=error_msg)
