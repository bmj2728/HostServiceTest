package hostserve

import (
	"fmt"
	"strings"
	"sync"
)

// HostServices provides functionalities for interacting with the host file system and environment variables.
type HostServices struct {
	ActiveClients *ActiveClients
	IHostFS
	IHostEnv
}

// NewHostServices creates a new HostServices instance using the provided file system and environment abstractions.
func NewHostServices(fs IHostFS, env IHostEnv) *HostServices {
	return &HostServices{
		ActiveClients: NewActiveClients(),
		IHostFS:       fs,
		IHostEnv:      env,
	}
}

type ActiveClientsMap map[ClientID]string

type ActiveClients struct {
	Clients ActiveClientsMap
	mu      sync.RWMutex
}

func NewActiveClients() *ActiveClients {
	return &ActiveClients{
		Clients: make(ActiveClientsMap),
	}
}

func (ac *ActiveClients) AddClient(client ClientID, owner string) error {
	if ac == nil {
		return fmt.Errorf("active clients map is nil")
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, exists := ac.Clients[client]; exists {
		return fmt.Errorf("client %s already exists", client)
	}

	ac.Clients[client] = owner
	return nil
}

func (ac *ActiveClients) RemoveClient(client ClientID) error {
	if ac == nil {
		return fmt.Errorf("no active clients")
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if _, exists := ac.Clients[client]; !exists {
		return fmt.Errorf("client %s does not exist", client)
	}
	delete(ac.Clients, client)
	return nil
}

func (ac *ActiveClients) GetClients() ActiveClientsMap {
	if ac == nil {
		return nil
	}
	return ac.Clients
}

func (ac *ActiveClients) GetClientOwner(client ClientID) (string, bool) {
	if ac == nil {
		return "", false
	}
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	owner, exists := ac.Clients[client]
	return owner, exists
}

func (ac *ActiveClients) FindClientsByOwner(owner string) ActiveClientsMap {
	if ac == nil {
		return nil
	}
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	clients := make(ActiveClientsMap)
	for client, clientOwner := range ac.Clients {
		if clientOwner == owner {
			clients[client] = clientOwner
		}
	}
	return clients
}

func (ac *ActiveClients) Len() int {
	if ac == nil {
		return 0
	}
	return len(ac.Clients)
}

func formatActiveClients(ac *ActiveClients) string {
	if ac == nil {
		return ""
	}
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	var b strings.Builder
	for client, owner := range ac.Clients {
		entry := fmt.Sprintf("%s: %s\n", client, owner)
		b.WriteString(entry)
	}
	return b.String()
}

func (ac *ActiveClients) String() string {
	return formatActiveClients(ac)
}

func (ac *ActiveClients) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.Clients = make(ActiveClientsMap)
}
