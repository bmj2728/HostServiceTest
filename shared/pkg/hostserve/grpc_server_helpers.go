package hostserve

import "fmt"

// getHostServices retrieves the HostServices instance from the server implementation or returns an error if invalid.
func (s *HostServiceGRPCServer) getHostServices() (*HostServices, error) {
	hs, ok := s.Impl.(*HostServices)
	if !ok {
		return nil, ErrInvalidHostServices
	}
	return hs, nil
}

// clientIdToOwner retrieves the owner of the client with the specified ID from the HostServices instance.
func (s *HostServiceGRPCServer) clientIdToOwner(id ClientID) (string, error) {
	hs, err := s.getHostServices()
	if err != nil {
		return "", err
	}
	owner, exists := hs.ActiveClients().GetClientOwner(id)
	if !exists {
		return "", fmt.Errorf("client %s is not registered", id)
	}
	return owner, nil
}
