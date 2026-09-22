package codexdriver

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
)

// Transport owns message movement only. The Codex driver owns JSON-RPC meaning.
type Transport interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
	Close() error
}

// JSONLTransport implements the default codex app-server stdio framing without
// owning process creation or shutdown policy.
type JSONLTransport struct {
	reader *bufio.Reader
	writer io.Writer
	closer io.Closer
	mu     sync.Mutex
}

func NewJSONLTransport(reader io.Reader, writer io.Writer, closer io.Closer) (*JSONLTransport, error) {
	if reader == nil || writer == nil {
		return nil, errors.New("codex transport requires reader and writer")
	}
	return &JSONLTransport{reader: bufio.NewReader(reader), writer: writer, closer: closer}, nil
}

func (t *JSONLTransport) Send(ctx context.Context, message []byte) error {
	if ctx == nil {
		return errors.New("codex transport requires context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(bytes.TrimSpace(message)) == 0 {
		return errors.New("codex transport refuses empty message")
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if _, err := t.writer.Write(append(append([]byte(nil), message...), '\n')); err != nil {
		return err
	}
	if flusher, ok := t.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

func (t *JSONLTransport) Receive(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		return nil, errors.New("codex transport requires context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil, errors.New("codex transport received empty frame")
	}
	return append([]byte(nil), line...), nil
}

func (t *JSONLTransport) Close() error {
	if t.closer == nil {
		return nil
	}
	return t.closer.Close()
}
