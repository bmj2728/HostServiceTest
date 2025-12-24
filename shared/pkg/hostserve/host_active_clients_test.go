package hostserve

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewActiveClients(t *testing.T) {
	ac := newActiveClients()
	if ac == nil {
		t.Fatal("Expected non-nil ActiveClients")
	}
	if ac.Clients == nil {
		t.Error("Expected initialized Clients map")
	}
	if len(ac.Clients) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(ac.Clients))
	}
}

func TestActiveClients_AddClient(t *testing.T) {
	ac := newActiveClients()

	t.Run("add new client", func(t *testing.T) {
		err := ac.AddClient("client-1", "owner-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(ac.Clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(ac.Clients))
		}
	})

	t.Run("add duplicate client returns error", func(t *testing.T) {
		err := ac.AddClient("client-1", "owner-1")
		if err == nil {
			t.Error("Expected error for duplicate client")
		}
	})

	t.Run("nil ActiveClients", func(t *testing.T) {
		var nilAC *ActiveClients
		err := nilAC.AddClient("test", "owner")
		if err == nil {
			t.Error("Expected error for nil ActiveClients")
		}
	})
}

func TestActiveClients_RemoveClient(t *testing.T) {
	ac := newActiveClients()
	ac.AddClient("client-1", "owner-1")

	t.Run("remove existing client", func(t *testing.T) {
		err := ac.RemoveClient("client-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(ac.Clients) != 0 {
			t.Errorf("Expected empty map after removal, got %d entries", len(ac.Clients))
		}
	})

	t.Run("remove non-existent client", func(t *testing.T) {
		err := ac.RemoveClient("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent client")
		}
	})

	t.Run("nil ActiveClients", func(t *testing.T) {
		var nilAC *ActiveClients
		err := nilAC.RemoveClient("test")
		if err == nil {
			t.Error("Expected error for nil ActiveClients")
		}
	})
}

func TestActiveClients_GetClientOwner(t *testing.T) {
	ac := newActiveClients()
	ac.AddClient("client-1", "owner-1")

	t.Run("get existing client owner", func(t *testing.T) {
		owner, exists := ac.GetClientOwner("client-1")
		if !exists {
			t.Error("Expected client to exist")
		}
		if owner != "owner-1" {
			t.Errorf("Expected owner 'owner-1', got %q", owner)
		}
	})

	t.Run("get non-existent client", func(t *testing.T) {
		_, exists := ac.GetClientOwner("non-existent")
		if exists {
			t.Error("Expected client to not exist")
		}
	})

	t.Run("nil ActiveClients", func(t *testing.T) {
		var nilAC *ActiveClients
		_, exists := nilAC.GetClientOwner("test")
		if exists {
			t.Error("Expected false for nil ActiveClients")
		}
	})
}

func TestActiveClients_FindClientsByOwner(t *testing.T) {
	ac := newActiveClients()
	ac.AddClient("client-1", "owner-1")
	ac.AddClient("client-2", "owner-1")
	ac.AddClient("client-3", "owner-2")

	t.Run("find multiple clients for owner", func(t *testing.T) {
		clients := ac.FindClientsByOwner("owner-1")
		if len(clients) != 2 {
			t.Errorf("Expected 2 clients, got %d", len(clients))
		}
	})

	t.Run("find single client for owner", func(t *testing.T) {
		clients := ac.FindClientsByOwner("owner-2")
		if len(clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(clients))
		}
	})

	t.Run("find no clients for owner", func(t *testing.T) {
		clients := ac.FindClientsByOwner("non-existent-owner")
		if len(clients) != 0 {
			t.Errorf("Expected 0 clients, got %d", len(clients))
		}
	})

	t.Run("nil ActiveClients", func(t *testing.T) {
		var nilAC *ActiveClients
		clients := nilAC.FindClientsByOwner("owner")
		if clients != nil {
			t.Error("Expected nil for nil ActiveClients")
		}
	})
}

func TestActiveClients_Len(t *testing.T) {
	ac := newActiveClients()

	if ac.Len() != 0 {
		t.Errorf("Expected len 0, got %d", ac.Len())
	}

	ac.AddClient("client-1", "owner-1")
	if ac.Len() != 1 {
		t.Errorf("Expected len 1, got %d", ac.Len())
	}

	ac.AddClient("client-2", "owner-2")
	if ac.Len() != 2 {
		t.Errorf("Expected len 2, got %d", ac.Len())
	}

	var nilAC *ActiveClients
	if nilAC.Len() != 0 {
		t.Error("Expected len 0 for nil ActiveClients")
	}
}

func TestActiveClients_Clear(t *testing.T) {
	ac := newActiveClients()
	ac.AddClient("client-1", "owner-1")
	ac.AddClient("client-2", "owner-2")

	ac.Clear()

	if ac.Len() != 0 {
		t.Errorf("Expected len 0 after Clear, got %d", ac.Len())
	}
}

func TestActiveClients_String(t *testing.T) {
	ac := newActiveClients()
	ac.AddClient("client-1", "owner-1")

	str := ac.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}

	var nilAC *ActiveClients
	if nilAC.String() != "" {
		t.Error("Expected empty string for nil ActiveClients")
	}
}

func TestActiveClients_ConcurrentAccess(t *testing.T) {
	ac := newActiveClients()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := ClientID(fmt.Sprintf("client-%d", id))
			ac.AddClient(clientID, fmt.Sprintf("owner-%d", id%10))
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ac.GetClients()
			_ = ac.Len()
		}()
	}
	wg.Wait()

	if ac.Len() == 0 {
		t.Error("Expected clients to be added")
	}
}

func TestClientID_String(t *testing.T) {
	id := ClientID("test-id")
	if id.String() != "test-id" {
		t.Errorf("Expected 'test-id', got %q", id.String())
	}
}

func TestNewClientID(t *testing.T) {
	id1 := newClientID()
	id2 := newClientID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id2 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}

	// UUID v7 or v4 format (36 chars with hyphens)
	if len(id1.String()) != 36 {
		t.Errorf("Expected UUID format (36 chars), got %d", len(id1.String()))
	}
}
