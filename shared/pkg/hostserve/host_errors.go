package hostserve

// HostServiceError represents an error returned by the host service.
// Message is a description of the error.
type HostServiceError struct {
	Message string
}

// Error returns the error message stored in the HostServiceError as a string.
func (e *HostServiceError) Error() string {
	return e.Message
}
