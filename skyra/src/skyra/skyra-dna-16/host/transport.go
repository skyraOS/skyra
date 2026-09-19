package host

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"sync"
)

// JSONLTransport adapts a newline-delimited byte stream to opaque complete
// frames. Protocol interpretation remains with its owner.
type JSONLTransport struct {
	reader *bufio.Reader
	writer io.Writer
	closer io.Closer
	mu     sync.Mutex
	once   sync.Once
}

func NewJSONLTransport(reader io.Reader, writer io.Writer, closer io.Closer) (*JSONLTransport, error) {
	if reader == nil || writer == nil {
		return nil, errors.New("transport requires reader and writer")
	}
	return &JSONLTransport{reader: bufio.NewReader(reader), writer: writer, closer: closer}, nil
}

func (transport *JSONLTransport) Send(ctx context.Context, frame []byte) error {
	if ctx == nil {
		return errors.New("transport requires context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	frame = bytes.TrimSpace(frame)
	if len(frame) == 0 {
		return errors.New("transport refuses empty frame")
	}
	transport.mu.Lock()
	defer transport.mu.Unlock()
	if _, err := transport.writer.Write(append(append([]byte(nil), frame...), '\n')); err != nil {
		return err
	}
	if flusher, ok := transport.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

func (transport *JSONLTransport) Receive(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		return nil, errors.New("transport requires context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	frame, err := transport.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	frame = bytes.TrimSpace(frame)
	if len(frame) == 0 {
		return nil, errors.New("transport received empty frame")
	}
	return append([]byte(nil), frame...), nil
}

func (transport *JSONLTransport) Close() error {
	var err error
	transport.once.Do(func() {
		if transport.closer != nil {
			err = transport.closer.Close()
		}
	})
	return err
}

var _ driver.Transport = (*JSONLTransport)(nil)
