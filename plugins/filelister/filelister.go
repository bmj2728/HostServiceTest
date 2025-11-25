package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type FileLister struct {
	fileHandles       map[string]hostserve.FileHandle
	broker            *plugin.GRPCBroker
	hostServiceClient hostserve.IHostServices
	conn              *grpc.ClientConn
	connMutex         sync.Mutex
}

func (f *FileLister) ListFiles(dir string) ([]string, error) {
	ctx := context.Background() //this is needed for the host service calls

	// simple env check
	home, err := f.hostServiceClient.GetEnv(ctx, "HOME")
	if err != nil {
		hclog.Default().Error("Failed to get env variable", "err", err)
	}

	// Read Dir
	dirEntries, err := f.hostServiceClient.ReadDir(ctx, dir)
	if err != nil {
		hclog.Default().Error("Failed to read directory via host service", "dir", dir, "err", err)
		return nil, err
	}

	var entries []string
	var buf bytes.Buffer
	entries = append(entries, home)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, entry.Name())
			buf.WriteString(entry.Name())
		} else {
			entries = append(entries, entry.Name())
			buf.WriteString(entry.Name())
		}
	}

	// Write file
	err = f.hostServiceClient.WriteFile(ctx, filepath.Join(dir, "listed_files.txt"), buf.Bytes(), 0644)
	if err != nil {
		hclog.Default().Error("Failed to write file via host service", "dir", dir, "err", err)
	}

	// Open file
	fh, sz, err := f.hostServiceClient.FileOpen(ctx, filepath.Join(dir, "listed_files.txt"), os.O_RDONLY, 0644)
	if err != nil {
		hclog.Default().Error("Failed to open file via host service", "dir", dir, "err", err)
	}
	hclog.Default().Info("Opened file", "handle", fh, "size", sz)

	// Close File called in a closure for deferment
	defer func(hostServiceClient hostserve.IHostServices, ctx context.Context, handle hostserve.FileHandle) {
		err := hostServiceClient.FileClose(ctx, handle)
		if err != nil {
			hclog.Default().Error("Failed to close file handle", "err", err)
		}
	}(f.hostServiceClient, ctx, fh)

	//store the file handle
	f.fileHandles[filepath.Join(dir, "listed_files.txt")] = fh

	stream, err := f.hostServiceClient.FileRead(ctx, fh, 64*1024)
	if err != nil {
		hclog.Default().Error("FileRead failed to read file via host service", "dir", dir, "err", err)
	}
	sb := make([]byte, 10)
	for {
		n, err := stream.Read(sb)
		if err != nil && err != io.EOF {
			hclog.Default().Error("FileRead failed to read file via host service", "dir", dir, "err", err)
			break
		}
		if n == 0 && err == io.EOF {
			hclog.Default().Info("FileRead reached end of file")
			break
		}
		hclog.Default().Info("FileRead read file", "read", n)
	}

	newFileName := filepath.Join(dir, "testing_open_create.txt")

	fh2, sz2, err := f.hostServiceClient.FileOpen(ctx, newFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		hclog.Default().Error("Failed to open file via host service", "dir", dir, "err", err)
	}
	hclog.Default().Info("Opened file", "handle", fh2, "size", sz2)

	f.fileHandles[newFileName] = fh2

	//retrieve the file handle
	retrieved, ok := f.fileHandles[newFileName]
	if !ok {
		hclog.Default().Error("Failed to retrieve file handle")
	}
	hclog.Default().Info("Retrieved file handle", "handle", retrieved)

	// just closing the file handle
	err = f.hostServiceClient.FileClose(ctx, retrieved)
	if err != nil {
		hclog.Default().Error("Failed to close file handle", "err", err)
	}

	return entries, nil
}

func (f *FileLister) EstablishHostServices(hostServiceID uint32) (hostserve.ClientID, error) {
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	conn, err := f.broker.Dial(hostServiceID)
	if err != nil {
		hclog.Default().Error("Failed to dial host service", "err", err)
		return "", fmt.Errorf("failed to dial broker: %w", err)
	}

	f.conn = conn
	client := hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))
	f.hostServiceClient = client
	return client.ClientID(), nil
}

func (f *FileLister) DisconnectHostServices() {
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	if f.conn != nil {
		if err := f.conn.Close(); err != nil {
			hclog.Default().Error("Failed to close connection", "err", err)
		}
		f.conn = nil
		f.hostServiceClient = nil
		hclog.Default().Info("Disconnected from host services")
	}
}

func (f *FileLister) SetBroker(broker *plugin.GRPCBroker) {
	f.broker = broker
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_KEY",
	MagicCookieValue: "TEST_VALUE",
}

func main() {
	fl := &FileLister{
		fileHandles: make(map[string]hostserve.FileHandle),
	}

	pluginMap := map[string]plugin.Plugin{
		"fl-plugin": &filelister.FileListerGRPCPlugin{Impl: fl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
