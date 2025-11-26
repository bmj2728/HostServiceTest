package hostserve

import (
	"io"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"google.golang.org/grpc"
)

// grpcFileStreamReader is a gRPC-based reader for streaming file data from a server, implementing
// the io.Reader interface.
// It buffers data from the stream and tracks the end of the data using the final field.
// stream represents the gRPC server stream for receiving file data chunks.
// buffer temporarily holds data read from the stream that has not yet been returned to the caller.
// final is a flag indicating whether the last data chunk was the final one in the stream.
type grpcFileStreamReader struct {
	stream grpc.ServerStreamingClient[hostservev1.FileReadResponse]
	buffer []byte
	final  bool
}

// Read reads data into the provided slice from the gRPC stream, returning the number of bytes read
// and any error encountered.
// If buffer data exists, it is read first; if the buffer is empty and the stream ends, it returns io.EOF.
// Any errors from the gRPC stream or response are wrapped in HostServiceError and returned.
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type grpcFileStreamWriter struct {
	stream grpc.ClientStreamingClient[hostservev1.FileWriteRequest, hostservev1.FileWriteResponse]
	handle FileHandle
	offset uint64
}

func (w *grpcFileStreamWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	totalWritten := 0
	data := p

	// Send data in chunks no larger than maxChunkSize
	for len(data) > 0 {
		chunkSize := len(data)
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		err := w.sendChunk(data[:chunkSize], false)
		if err != nil {
			return totalWritten, err
		}

		data = data[chunkSize:]
		totalWritten += chunkSize
	}

	return totalWritten, nil
}

func (w *grpcFileStreamWriter) sendChunk(data []byte, isFinal bool) error {
	err := w.stream.Send(&hostservev1.FileWriteRequest{
		Handle: w.handle.String(),
		Chunk: &hostservev1.FileChunk{
			Data:    data,
			Offset:  w.offset,
			IsFinal: isFinal,
		},
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	w.offset += uint64(len(data))
	return nil
}

func (w *grpcFileStreamWriter) Close() error {
	// Send final marker
	err := w.sendChunk(nil, true)
	if err != nil {
		return err
	}

	// Receive the final response
	resp, err := w.stream.CloseAndRecv()
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}

	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}

	return nil
}
