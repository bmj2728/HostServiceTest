package main

//note that we do not need to import os or fs here, as we are using the host service to read the files
import (
	"sync"

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

func (f *ColorLister) ListFiles(dir string) ([]string, error) {
	//uses host to read dir vs. using os.ReadDir(dir) or fs.ReadDir(fs, dir)
	dirEntries, err := f.hostServiceClient.ReadDir(dir)
	if err != nil {
		hclog.Default().Error("Failed to read directory via host service", "dir", dir, "err", err)
		return nil, err
	}

	var entries []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, dirFormat.Wrap(entry.Name()+"-d", true))
		} else {
			data, err := f.hostServiceClient.ReadFile(dir, entry.Name())
			if err != nil {
				hclog.Default().Error("Failed to read file via host service", "dir", dir,
					"file", entry.Name(), "err", err)
			}
			contents := string(data)
			entries = append(entries, fileFormat.Wrap(entry.Name()+"-f", true))
			entries = append(entries, "Contents:\n", contents)
		}
	}

	return entries, nil
}

func (f *ColorLister) EstablishHostServices(hostServiceID uint32) {
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	conn, err := f.broker.Dial(hostServiceID)
	if err != nil {
		hclog.Default().Error("Failed to dial host service", "err", err)
		return
	}

	f.conn = conn
	f.hostServiceClient = hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))
	hclog.Default().Info("Established host services", "id", hostServiceID)
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
