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

// buildLogFields constructs the base logging fields and appends request-specific fields.
// Returns a slice of key-value pairs suitable for hclog.
func buildLogFields(clientID ClientID, owner string, reqID RequestID, request proto.Message) []interface{} {
	baseFields := []interface{}{
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	}

	// Extract request fields if we have a valid proto message
	if request != nil {
		requestFields := extractRequestFields("", request, 0)
		return append(baseFields, requestFields...)
	}

	return baseFields
}

// recordMetrics updates the metrics with end time and duration.
func recordMetrics(metrics *RequestMetrics) {
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
}

// logBadRequest logs a context validation error with timing information.
func logBadRequest(operationName string, logFields []interface{}, err error, metrics *RequestMetrics) {
	recordMetrics(metrics)
	metrics.Success = false

	hclog.Default().Info(fmt.Sprintf("%s bad request from client", operationName),
		append(logFields,
			"error", err,
			"duration_us", metrics.Duration.Microseconds(),
			"success", false,
		)...)
}

// logRequestStart logs the beginning of request processing.
func logRequestStart(operationName string, logFields []interface{}) {
	hclog.Default().Info(fmt.Sprintf("%s request from client", operationName), logFields...)
}

// logCompletion logs the completion of a request with metrics and outcome.
func logCompletion(operationName string, logFields []interface{}, metrics *RequestMetrics, err error) {
	completionFields := append(logFields,
		"duration_us", metrics.Duration.Microseconds(),
		"end_time", metrics.EndTime.Format(time.RFC3339Nano),
		"success", metrics.Success,
	)

	if err != nil {
		completionFields = append(completionFields, "error", err)
		hclog.Default().Warn(fmt.Sprintf("%s failed", operationName), completionFields...)
	} else if !metrics.Success {
		hclog.Default().Warn(fmt.Sprintf("%s failed", operationName), completionFields...)
	} else {
		hclog.Default().Info(fmt.Sprintf("%s completed successfully", operationName), completionFields...)
	}
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
	var zeroValue Res
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Process and validate context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	logFields := buildLogFields(clientID, owner, reqID, request)

	// Handle validation errors
	if err != nil {
		logBadRequest(operationName, logFields, err, metrics)
		return zeroValue, err
	}

	// Log request start
	logRequestStart(operationName, logFields)

	// Execute handler
	result, handlerErr := handler(ctx, request)

	// Record metrics
	recordMetrics(metrics)
	metrics.Success = handlerErr == nil && !hasErrorField(result)

	// Extract error message from response if present
	var errMsg error = handlerErr
	if !metrics.Success && handlerErr == nil {
		if msg := getErrorField(result); msg != "" {
			errMsg = fmt.Errorf("%s", msg)
		}
	}

	// Log completion
	logCompletion(operationName, logFields, metrics, errMsg)

	// Return result
	if handlerErr != nil {
		return zeroValue, handlerErr
	}
	return result, nil
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
	ctx := stream.Context()
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Process and validate context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	logFields := buildLogFields(clientID, owner, reqID, request)

	// Handle context validation errors - missing client id/owner or request id
	if err != nil {
		logBadRequest(operationName, logFields, err, metrics)
		return err
	}

	// Log request start
	logRequestStart(operationName, logFields)

	// Execute handler - See grpc_server_fs/FileReader for an example closure containing the logic for the operation
	// handler should encompass all logic to process the request
	handlerErr := handler(ctx, clientID, reqID, owner, logFields)

	// Record metrics
	recordMetrics(metrics)
	metrics.Success = handlerErr == nil

	// Log completion
	logCompletion(operationName, logFields, metrics, handlerErr)

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
	ctx := stream.Context()
	metrics := &RequestMetrics{StartTime: time.Now()}

	// Receive first message to get request metadata
	firstReq, err := stream.Recv()
	if err != nil {
		return err
	}

	// Process and validate context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)

	// Build logging fields - extract from the first message if it's a proto.Message
	var protoMsg proto.Message
	if msg, ok := any(firstReq).(proto.Message); ok {
		protoMsg = msg
	}
	logFields := buildLogFields(clientID, owner, reqID, protoMsg)

	// Handle validation errors
	if err != nil {
		logBadRequest(operationName, logFields, err, metrics)
		return err
	}

	// Log request start
	logRequestStart(operationName, logFields)

	// Execute handler
	handlerErr := handler(ctx, clientID, reqID, owner, logFields, firstReq, stream)

	// Record metrics
	recordMetrics(metrics)
	metrics.Success = handlerErr == nil

	// Log completion
	logCompletion(operationName, logFields, metrics, handlerErr)

	return handlerErr
}
