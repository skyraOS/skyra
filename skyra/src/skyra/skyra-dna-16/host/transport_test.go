package host

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

func TestJSONLTransportFramesConcurrentMessages(t *testing.T) {
	var stream bytes.Buffer
	transport, err := NewJSONLTransport(&stream, &stream, nil)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for _, frame := range []string{`{"id":1}`, `{"id":2}`, `{"id":3}`} {
		frame := frame
		wait.Add(1)
		go func() { defer wait.Done(); _ = transport.Send(context.Background(), []byte(frame)) }()
	}
	wait.Wait()
	for index := 0; index < 3; index++ {
		frame, err := transport.Receive(context.Background())
		if err != nil || !bytes.HasPrefix(frame, []byte(`{"id":`)) {
			t.Fatalf("invalid frame %q: %v", frame, err)
		}
	}
}

func TestTransportCloseUnblocksReceive(t *testing.T) {
	reader, writer := io.Pipe()
	transport, err := NewJSONLTransport(reader, io.Discard, reader)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := transport.Receive(context.Background()); done <- err }()
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("receive succeeded after close")
		}
	case <-time.After(time.Second):
		t.Fatal("close did not unblock receive")
	}
}

func TestProcessHostRoundTripAndClose(t *testing.T) {
	if _, err := os.Stat("/bin/cat"); err != nil {
		t.Skip("/bin/cat unavailable")
	}
	host, err := StartProcessHost(context.Background(), ProcessSpec{Path: "/bin/cat", Env: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Transport().Send(context.Background(), []byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	frame, err := host.Transport().Receive(context.Background())
	if err != nil || string(frame) != `{"ok":true}` {
		t.Fatalf("round trip %q: %v", frame, err)
	}
	if err := host.Close(); err != nil && err != os.ErrProcessDone {
		t.Fatal(err)
	}
	select {
	case <-host.Done():
	case <-time.After(time.Second):
		t.Fatal("process did not stop")
	}
}
