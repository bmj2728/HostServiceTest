package hostserve

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ctxClientIDKey is the context key used to store the client identifier in a context for outgoing requests.
const ctxClientIDKey = "client"
const ctxHostRequestIDKey = "request"
const ctxClientOwner = "clientOwner"

// addTracingIDsToContext appends client ID and request ID metadata to an outgoing context for tracing in gRPC calls.
func addTracingIDsToContext(ctx context.Context, clientID ClientID, requestID RequestID) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		ctxClientIDKey, clientID.String(),
		ctxHostRequestIDKey, requestID.String())
}

// addClientOwnerToContext adds the client owner information to the given context and returns the updated context.
func addClientOwnerToContext(ctx context.Context, owner string) context.Context {
	return context.WithValue(ctx, ctxClientOwner, owner)
}

// getClientIDFromContext extracts the client ID from the provided gRPC context and returns it as a string.
// Returns an empty string if no client ID is found or the metadata is unavailable.
func getClientIDFromContext(ctx context.Context) ClientID {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	clientID := md.Get(ctxClientIDKey)
	if len(clientID) == 0 {
		return ""
	}
	return ClientID(clientID[0])
}

// getRequestIDFromContext extracts the request ID from the context's metadata and returns it as a RequestID.
// If the request ID is missing, an empty string is returned.
func getRequestIDFromContext(ctx context.Context) RequestID {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	requestID := md.Get(ctxHostRequestIDKey)
	if len(requestID) == 0 {
		return ""
	}
	return RequestID(requestID[0])
}

// getClientOwnerFromContext retrieves the client owner information from the given context.
// If the context does not contain the client owner, an empty string is returned.
func getClientOwnerFromContext(ctx context.Context) string {
	value := ctx.Value(ctxClientOwner)
	if value == nil {
		return ""
	}
	return value.(string)
}
