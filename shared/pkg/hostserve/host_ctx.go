package hostserve

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ctxClientIDKey is the context key used to store the client identifier in a context for outgoing requests.
const ctxClientIDKey = "client"
const ctxHostRequestIDKey = "request"

func addTracingIDsToContext(ctx context.Context, clientID ClientID, requestID RequestID) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		ctxClientIDKey, clientID.String(),
		ctxHostRequestIDKey, requestID.String())
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
