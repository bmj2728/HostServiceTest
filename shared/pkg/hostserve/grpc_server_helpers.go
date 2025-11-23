package hostserve

import (
	"context"
	"fmt"
)

var (
	ErrEmptyClientID  = fmt.Errorf("client ID cannot be empty")
	ErrEmptyRequestID = fmt.Errorf("request ID cannot be empty")
)

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

// processRequestContext extracts client and request information from context,
// adds client owner, and validates required fields.
func (s *HostServiceGRPCServer) processRequestContext(ctx context.Context) (context.Context, ClientID, RequestID, string, error) {

	clientID := getClientIDFromContext(ctx)
	if clientID == "" {
		return nil, "", "", "", ErrEmptyClientID
	}
	reqID := getRequestIDFromContext(ctx)
	if reqID == "" {
		return nil, "clientID", "", "", ErrEmptyRequestID
	}
	owner, err := s.clientIdToOwner(clientID)
	if err != nil {
		return ctx, clientID, reqID, "", fmt.Errorf("failed to get client owner: %w", err)
	}
	ctx = addClientOwnerToContext(ctx, owner)
	return ctx, clientID, reqID, owner, nil
}
