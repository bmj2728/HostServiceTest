# -*- mode: python ; coding: utf-8 -*-

"""
PyInstaller spec file for pylelister plugin.

This bundles the Python interpreter, plugin code, and all dependencies
into a single executable binary that can be run without any Python
installation on the host system.
"""

import os
import sys

# Get the absolute path to the project root
project_root = os.path.abspath(os.path.join(SPECPATH, '..', '..'))
protogen_py = os.path.join(project_root, 'shared', 'protogen_py')

a = Analysis(
    ['src/main.py', 'src/plugin.py'],  # Include both main and plugin modules
    pathex=[
        SPECPATH,
        protogen_py,  # Include generated protobuf code
        os.path.join(SPECPATH, 'src'),  # Include src directory
    ],
    binaries=[],
    datas=[],
    hiddenimports=[
        'grpc',
        'google.protobuf',
        'concurrent.futures',
        'uuid',
        # Generated protobuf modules
        'filelister.v1.filelister_pb2',
        'filelister.v1.filelister_pb2_grpc',
        'hostserve.v1.hostserve_pb2',
        'hostserve.v1.hostserve_pb2_grpc',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=None,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=None)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='pylelister',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,  # Keep console for stdio communication with go-plugin
    disable_windowed_traceback=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
