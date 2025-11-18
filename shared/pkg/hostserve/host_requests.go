package hostserve

// RequestID represents a unique identifier for a specific request, typically used for tracing and tracking purposes.
type RequestID string

func NewRequestID() RequestID {
	return RequestID(newUUID().String())
}

// String converts the RequestID value to its string representation.
func (rid RequestID) String() string {
	return string(rid)
}
