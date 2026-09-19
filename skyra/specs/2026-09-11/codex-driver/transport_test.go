package codexdriver

import (
	"bytes"
	"context"
	"testing"
)

func TestJSONLTransportFramesMessages(t *testing.T) {
	var stream bytes.Buffer
	transport, err := NewJSONLTransport(&stream, &stream, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := transport.Send(context.Background(), []byte(`{"id":1}`)); err != nil {
		t.Fatal(err)
	}
	message, err := transport.Receive(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(message) != `{"id":1}` {
		t.Fatalf("message = %q", message)
	}
}
