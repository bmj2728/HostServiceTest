package main

//note that we do not need to import os or fs here, as we are using the host service to read the files
import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/novelgitllc/ansicolor/v3"
	"google.golang.org/grpc"
)

var (
	fileFormat = ansicolor.NewFormat().WithForeground(ansicolor.FgBrightBlue)
	dirFormat  = ansicolor.NewFormat().WithForeground(ansicolor.FgBrightGreen)
)

type ColorLister struct {
	broker            *plugin.GRPCBroker
	hostServiceClient hostserve.IHostServices
	conn              *grpc.ClientConn
	connMutex         sync.Mutex
}

func (f *ColorLister) ListFiles(rootDir, path string) ([]string, error) {
	ctx := context.Background()

	dirEntries, err := f.hostServiceClient.ReadDir(ctx, rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to read directory via host service", "root", rootDir, "path", path, "err", err)
		return nil, err
	}

	var entries []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, dirFormat.Wrap(entry.Name(), true))
		} else {
			if entry.Name() == "renameme.md" {
				continue
			}
			data, err := f.hostServiceClient.ReadFile(ctx, rootDir, entry.Name())
			if err != nil {
				hclog.Default().Error("Failed to read file via host service", "root", rootDir, "path", path,
					"file", entry.Name(), "err", err)
				continue
			}
			contents := len(string(data))
			entries = append(entries, fileFormat.Wrap(entry.Name(), true)+fmt.Sprintf(" Size: %d bytes", contents))
		}
	}

	err = f.hostServiceClient.Mkdir(ctx, rootDir, "created_dir", 0755)
	if err != nil {
		hclog.Default().Error("Failed to create directory via host service", "root", rootDir, "err", err)
		return nil, err
	}

	err = f.hostServiceClient.MkdirAll(ctx, rootDir, "nested/dir", 0755)
	if err != nil {
		hclog.Default().Error("Failed to create directory via host service", "root", rootDir, "err", err)
		return nil, err
	}

	fh, err := f.hostServiceClient.FileCreate(ctx, rootDir, "nested/dir/created_file.txt")
	if err != nil {
		hclog.Default().Error("Failed to create file via host service", "dir", rootDir, "err", err)
		return nil, err
	}
	defer func(ctx context.Context, handle hostserve.FileHandle) {
		err := f.hostServiceClient.FileClose(ctx, handle)
		if err != nil {
			hclog.Default().Error("Failed to close file handle", "err", err)
		}
	}(ctx, fh)

	newOff, err := f.hostServiceClient.FileSeek(ctx, fh, 0, io.SeekStart)
	if err != nil {
		hclog.Default().Error("Failed to seek file via host service", "dir", rootDir, "err", err)
		return nil, err
	}
	hclog.Default().Info("File seeked", "offset", newOff)

	err = f.hostServiceClient.Chmod(ctx, rootDir, "mode_change.txt", 0644)
	if err != nil {
		hclog.Default().Error("Failed to change file mode", "err", err)
	}

	err = f.hostServiceClient.Chown(ctx, rootDir, "owner_change.txt", 1001, 1001)
	if err != nil {
		hclog.Default().Error("Failed to change file owner and group", "err", err)
	}

	err = f.hostServiceClient.Chtimes(ctx, rootDir, "owner_change.txt", time.Now(), time.Now())
	if err != nil {
		hclog.Default().Error("Failed to change file modification time", "err", err)
	}

	return entries, nil
}

func (f *ColorLister) EstablishHostServices(hostServiceID uint32) (hostserve.ClientID, error) {
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

func (f *ColorLister) DisconnectHostServices() {
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

func (f *ColorLister) SetBroker(broker *plugin.GRPCBroker) {
	f.broker = broker
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_KEY",
	MagicCookieValue: "TEST_VALUE",
}

func main() {
	cl := &ColorLister{}

	pluginMap := map[string]plugin.Plugin{
		"cl-plugin": &filelister.FileListerGRPCPlugin{Impl: cl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
