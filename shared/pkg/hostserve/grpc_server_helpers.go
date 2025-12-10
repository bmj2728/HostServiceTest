package hostserve

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	ErrEmptyClientID  = fmt.Errorf("client ID cannot be empty")
	ErrEmptyRequestID = fmt.Errorf("request ID cannot be empty")
)

// RequestMetrics tracks timing and success metrics for a single gRPC request.
// This struct captures the complete lifecycle of a request for logging and monitoring purposes.
type RequestMetrics struct {
	StartTime time.Time     // When the request handler was invoked
	EndTime   time.Time     // When the request handler completed (success or failure)
	Duration  time.Duration // Total time spent processing the request
	Success   bool          // Whether the request completed without errors
}

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

// hasErrorField checks if a protobuf message has a populated Error field.
// Returns true if the message has an Error field and it contains a non-empty string.
func hasErrorField(msg proto.Message) bool {
	if msg == nil {
		return false
	}
	return getErrorField(msg) != ""
}

// getErrorField extracts the error message from a protobuf response's Error field.
// Returns empty string if no Error field exists or it's nil/empty.
func getErrorField(msg proto.Message) string {
	if msg == nil {
		return ""
	}

	// Use protobuf reflection to find the "error" field
	msgReflect := msg.ProtoReflect()
	fields := msgReflect.Descriptor().Fields()

	// Look for a field named "error"
	errorField := fields.ByName("error")
	if errorField == nil {
		return ""
	}

	// Check if the field is set
	if !msgReflect.Has(errorField) {
		return ""
	}

	// Get the field value - it should be a string
	fieldValue := msgReflect.Get(errorField)
	if errorField.Kind() == 9 { // 9 is protoreflect.StringKind
		return fieldValue.String()
	}

	return ""
}

// withRequestLoggingAndResponse wraps standard unary gRPC request handlers with logging and metrics.
// This helper handles the complete request lifecycle for handlers that receive one request and return one response.
//
// Use this for standard unary RPC methods like: ReadDir, ReadFile, Chown, Mkdir, etc.
// DO NOT use this for streaming RPCs (FileReader, FileWriter) - those require specialized handling.
//
// The wrapper performs:
// - Context processing and validation (clientID, requestID, owner extraction)
// - Request logging with all protobuf fields extracted via reflection
// - Handler execution with timing
// - Completion logging with duration, end time, success/failure status, and error details
//
// Type Parameters:
//   - Req: The protobuf request type (must implement proto.Message)
//   - Res: The protobuf response type (must implement proto.Message)
//
// Parameters:
//   - s: The gRPC server instance
//   - ctx: The gRPC request context containing client metadata
//   - operationName: The operation name for logging (e.g., "Chown", "ReadDir")
//   - request: The protobuf request message
//   - handler: The business logic that processes the validated context and request
//
// Returns:
//   - Res: The response from the handler
//   - error: Context validation error or handler error
//
// Logs three types of entries:
// 1. "bad request" - context validation failed (missing clientID, unregistered client, etc.)
// 2. "request from client" - valid request started processing
// 3. "completed successfully" / "failed" - request finished with timing and outcome
//
// All log entries include:
//   - clientID, owner, requestID: Request traceability
//   - All non-zero request fields: Complete parameter visibility
//   - duration_ms: Processing time in milliseconds
//   - end_time: ISO8601 timestamp of completion
//   - success: Boolean outcome indicator
//   - error: Error message if applicable
func withRequestLoggingAndResponse[Req proto.Message, Res proto.Message](
	s *HostServiceGRPCServer,
	ctx context.Context,
	operationName string,
	request Req,
	handler func(context.Context, Req) (Res, error),
) (Res, error) {
	// Initialize zero value for early returns on error
	var zeroValue Res

	// Start timing the request
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Process and validate the request context.
	// Extracts clientID and requestID from gRPC metadata.
	// Validates both are present and non-empty.
	// Looks up client owner from active clients registry.
	// Adds owner to context for downstream handlers.
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)

	// Build base logging fields for request traceability
	baseFields := []interface{}{
		ctxClientIDKey, clientID, // Which client made this request
		ctxClientOwner, owner, // Owner/identity of the client
		ctxHostRequestIDKey, reqID, // Unique ID for this specific request
	}

	// Extract all non-zero fields from the protobuf request using reflection.
	// This provides complete visibility into request parameters without manual field enumeration.
	requestFields := extractRequestFields("", request, 0)

	// Combine base context fields with request-specific fields
	logFields := append(baseFields, requestFields...)

	// Handle context validation errors
	if err != nil {
		// Record timing for failed validation
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false

		// Log bad request with full context for debugging client issues.
		// Even though validation failed, we log whatever context we could extract.
		hclog.Default().Info(fmt.Sprintf("%s bad request from client", operationName),
			append(logFields,
				"error", err, // Why validation failed
				"duration_us", metrics.Duration.Microseconds(), // Time spent in validation
				"success", false, // Explicit failure marker
			)...)

		// Return zero value and validation error
		return zeroValue, err
	}

	// Log the validated request start.
	// Creates audit trail of all valid requests entering the system.
	hclog.Default().Info(fmt.Sprintf("%s request from client", operationName), logFields...)

	// Execute the business logic handler with validated context and request
	result, handlerErr := handler(ctx, request)

	// Record completion timing and outcome
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	// Determine success: check both Go error and protobuf Error field
	// Most handlers return errors in the response's Error field, not as Go errors
	metrics.Success = handlerErr == nil && !hasErrorField(result)

	// Build completion log with timing and outcome metrics.
	// Enables performance monitoring, success rate tracking, and error correlation.
	completionFields := append(logFields,
		"duration_us", metrics.Duration.Microseconds(), // Processing time in microseconds
		"end_time", metrics.EndTime.Format(time.RFC3339Nano), // Completion timestamp
		"success", metrics.Success, // Success/failure indicator
	)

	// Log completion with outcome-specific message
	if handlerErr != nil {
		// Add error details for failed requests (Go error path)
		completionFields = append(completionFields, "error", handlerErr)
		hclog.Default().Warn(fmt.Sprintf("%s failed", operationName), completionFields...)

		// Return zero value and handler error
		return zeroValue, handlerErr
	} else if !metrics.Success {
		// Handler succeeded at Go level but returned error in response message
		// Extract error from response for logging
		if errorMsg := getErrorField(result); errorMsg != "" {
			completionFields = append(completionFields, "error", errorMsg)
		}
		hclog.Default().Warn(fmt.Sprintf("%s failed", operationName), completionFields...)

		// Return handler result (which contains the error in its Error field)
		return result, nil
	} else {
		// Log successful completion
		hclog.Default().Info(fmt.Sprintf("%s completed successfully", operationName), completionFields...)

		// Return handler result
		return result, nil
	}
}

// withServerStreamLogging wraps server streaming gRPC handlers with logging and metrics.
// Server streaming RPCs receive one request and send multiple responses via a stream.
//
// Use this ONLY for server streaming methods like FileReader where we:
// - Receive a single request from the client
// - Return multiple streaming responses over time
// - Track total bytes/chunks sent
//
// This differs from unary RPCs because:
// - The handler doesn't return a response value, it streams via the stream parameter
// - Success/failure is determined by the error return and stream state
// - We track streaming metrics (chunks sent, bytes transferred)
//
// Parameters:
//   - stream: The gRPC server streaming server
//   - operationName: Operation name for logging (e.g., "FileReader")
//   - request: The protobuf request message
//   - handler: The streaming logic that sends responses via the stream
//
// Returns:
//   - error: Context validation error, handler error, or stream error
//
// The handler function receives:
//   - ctx: Validated context with client metadata
//   - clientID, reqID, owner: For additional logging within the handler
//   - logFields: Pre-built log fields for consistent logging in handler
//
// Logs:
// 1. "bad request" - context validation failed
// 2. "request from client" - stream started
// 3. "completed successfully" / "failed" - stream finished with metrics
func withServerStreamLogging[Req proto.Message, Res any](
	s *HostServiceGRPCServer,
	stream grpc.ServerStreamingServer[Res],
	operationName string,
	request Req,
	handler func(context.Context, ClientID, RequestID, string, []interface{}) error,
) error {
	// Get context from stream
	ctx := stream.Context()

	// Start timing
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Process and validate context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)

	// Build logging fields
	baseFields := []interface{}{
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	}
	requestFields := extractRequestFields("", request, 0)
	logFields := append(baseFields, requestFields...)

	// Handle context validation errors
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false

		// Log bad request
		hclog.Default().Error(fmt.Sprintf("%s bad request from client", operationName),
			append(logFields,
				"error", err,
				"duration_ms", metrics.Duration.Milliseconds(),
				"success", false,
			)...)

		return err
	}

	// Log stream start
	hclog.Default().Info(fmt.Sprintf("%s request from client", operationName), logFields...)

	// Execute streaming handler
	handlerErr := handler(ctx, clientID, reqID, owner, logFields)

	// Record completion
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = handlerErr == nil

	// Build completion fields
	completionFields := append(logFields,
		"duration_ms", metrics.Duration.Milliseconds(),
		"end_time", metrics.EndTime.Format(time.RFC3339Nano),
		"success", metrics.Success,
	)

	// Log completion
	if handlerErr != nil {
		completionFields = append(completionFields, "error", handlerErr)
		hclog.Default().Info(fmt.Sprintf("%s failed", operationName), completionFields...)
	} else {
		hclog.Default().Info(fmt.Sprintf("%s completed successfully", operationName), completionFields...)
	}

	return handlerErr
}

// withClientStreamLogging wraps client streaming gRPC handlers with logging and metrics.
// Client streaming RPCs receive multiple requests via a stream and return one response.
//
// Use this ONLY for client streaming methods like FileWriter where we:
// - Receive multiple streaming requests from the client
// - Process them incrementally
// - Return a single final response
//
// This differs from other patterns because:
// - The first message must be received to get the request metadata
// - Context processing happens after receiving the first message
// - We track streaming metrics (chunks received, bytes written)
//
// Parameters:
//   - stream: The gRPC client streaming server
//   - operationName: Operation name for logging (e.g., "FileWriter")
//   - handler: The streaming logic that receives from stream and processes
//
// Returns:
//   - error: Context validation error, handler error, or stream error
//
// The handler function receives:
//   - ctx: Validated context with client metadata
//   - clientID, reqID, owner: For logging within the handler
//   - logFields: Pre-built log fields for consistent logging
//   - firstReq: The first request message received (contains handle/metadata)
//
// Logs:
// 1. "bad request" - context validation failed after receiving first message
// 2. "request from client" - stream processing started
// 3. "completed successfully" / "failed" - stream finished with metrics
func withClientStreamLogging[Req any, Res any](
	s *HostServiceGRPCServer,
	stream grpc.ClientStreamingServer[Req, Res],
	operationName string,
	handler func(context.Context, ClientID, RequestID, string, []interface{}, *Req, grpc.ClientStreamingServer[Req, Res]) error,
) error {
	// Get context from stream
	ctx := stream.Context()

	// Start timing
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Receive first message to get request metadata
	// (we can't process context until we have the request)
	firstReq, err := stream.Recv()
	if err != nil {
		return err
	}

	// Process and validate context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)

	// Build logging fields
	// For client streaming, we log whatever we can extract from the first message
	baseFields := []interface{}{
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	}

	// Handle context validation errors
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false

		// Log bad request
		hclog.Default().Info(fmt.Sprintf("%s bad request from client", operationName),
			append(baseFields,
				"error", err,
				"duration_ms", metrics.Duration.Milliseconds(),
				"success", false,
			)...)

		return err
	}

	// Log stream start
	hclog.Default().Info(fmt.Sprintf("%s request from client", operationName), baseFields...)

	// Execute streaming handler, passing the first request and stream
	handlerErr := handler(ctx, clientID, reqID, owner, baseFields, firstReq, stream)

	// Record completion
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = handlerErr == nil

	// Build completion fields
	completionFields := append(baseFields,
		"duration_ms", metrics.Duration.Milliseconds(),
		"end_time", metrics.EndTime.Format(time.RFC3339Nano),
		"success", metrics.Success,
	)

	// Log completion
	if handlerErr != nil {
		completionFields = append(completionFields, "error", handlerErr)
		hclog.Default().Info(fmt.Sprintf("%s failed", operationName), completionFields...)
	} else {
		hclog.Default().Info(fmt.Sprintf("%s completed successfully", operationName), completionFields...)
	}

	return handlerErr
}
