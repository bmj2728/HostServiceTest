package hostserve

import (
	"io"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"google.golang.org/grpc"
)

type grpcFileStreamReader struct {
	stream grpc.ServerStreamingClient[hostservev1.FileReadResponse]
	buffer []byte
	final  bool // track if buffered data is the end
}

func (g *grpcFileStreamReader) Read(p []byte) (n int, err error) {

	// Check for zero-length buffer
	if len(p) == 0 {
		return 0, nil
	}

	// If we buffered data on the last read, add it to p first
	if len(g.buffer) > 0 {
		n = copy(p, g.buffer)
		g.buffer = g.buffer[n:]

		// When the buffer is empty and the last read reported EOF, return EOF
		if len(g.buffer) == 0 && g.final {
			return n, io.EOF
		}
		return n, nil
	}

	// Read from the stream
	resp, recvErr := g.stream.Recv()

	// Check if there's data in the response
	if resp != nil && resp.Chunk != nil && len(resp.Chunk.Data) > 0 {
		// Copy what fits into p
		n = copy(p, resp.Chunk.Data)

		// Buffer remaining data
		if n < len(resp.Chunk.Data) {
			g.buffer = resp.Chunk.Data[n:]
			// Track if this was the final chunk
			g.final = resp.Chunk.IsFinal
			return n, nil // Return data now, error (if any) on next Read()
		}

		// All data fit - check if this was the final chunk
		if resp.Chunk.IsFinal {
			return n, io.EOF
		}

		return n, nil
	}

	// No valid data in response, handle errors
	if recvErr == io.EOF {
		return 0, io.EOF
	}
	if recvErr != nil {
		return 0, &HostServiceError{Message: recvErr.Error()}
	}

	// Check for error in response message
	if resp != nil && resp.Error != nil {
		return 0, &HostServiceError{Message: *resp.Error}
	}

	// Got response but no data and no error - shouldn't happen
	return 0, &HostServiceError{Message: "empty response from stream"}
}
