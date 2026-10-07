package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

const token = "01234567890123456789012345678901"

func newServer(t *testing.T, call Call) *Server {
	t.Helper()
	s, err := New("frame-1", token, []Tool{{Name: "echo", InputSchema: json.RawMessage(`{"type":"object"}`)}}, call)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func invoke(s *Server, body, auth string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader(body))
	r.Header.Set("Authorization", auth)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestRecordedDeliveryDeduplication(t *testing.T) {
	var calls atomic.Int32
	s := newServer(t, func(ctx context.Context, id, name string, args json.RawMessage) (json.RawMessage, error) {
		calls.Add(1)
		if id != `frame-1:mcp:7` || name != "echo" || string(args) != `{"text":"hi"}` {
			t.Errorf("unexpected delivery %s %s %s", id, name, args)
		}
		return json.RawMessage(`{"text":"hi"}`), nil
	})
	body := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"echo","arguments":{"text":"hi"}}}`
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := invoke(s, body, "Bearer "+token)
			if !strings.Contains(w.Body.String(), `\"text\":\"hi\"`) {
				t.Errorf("%s", w.Body.String())
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("delivered %d times", calls.Load())
	}
	changed := strings.Replace(body, `"hi"`, `"other"`, 1)
	if !strings.Contains(invoke(s, changed, "Bearer "+token).Body.String(), `"isError":true`) {
		t.Fatal("accepted conflicting retry")
	}
	if calls.Load() != 1 {
		t.Fatal("conflicting retry executed")
	}
}

func TestCapabilityReachAndRedaction(t *testing.T) {
	var calls int
	s := newServer(t, func(context.Context, string, string, json.RawMessage) (json.RawMessage, error) {
		calls++
		return nil, errors.New("credential=do-not-expose")
	})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{}}}`
	if invoke(s, body, "").Code != http.StatusUnauthorized {
		t.Fatal("missing capability accepted")
	}
	outside := strings.Replace(body, `"echo"`, `"shell"`, 1)
	if !strings.Contains(invoke(s, outside, "Bearer "+token).Body.String(), "outside reach") || calls != 0 {
		t.Fatal("out-of-reach tool delivered")
	}
	w := invoke(s, body, "Bearer "+token)
	if strings.Contains(w.Body.String(), "do-not-expose") || !strings.Contains(w.Body.String(), `"isError":true`) {
		t.Fatal(w.Body.String())
	}
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Origin", "https://untrusted.invalid")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatal("browser origin accepted")
	}
}

func TestProtocol(t *testing.T) {
	s := newServer(t, func(context.Context, string, string, json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{}`), nil
	})
	for _, tc := range []struct{ body, want string }{
		{`{"jsonrpc":"2.0","id":"i","method":"initialize","params":{"protocolVersion":"2025-03-26"}}`, `"protocolVersion":"2025-03-26"`},
		{`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, `"name":"echo"`},
		{`{"jsonrpc":"2.0","id":3,"method":"ping"}`, `"result":{}`},
		{`{"jsonrpc":"2.0","id":4,"method":"execute"}`, `"code":-32601`},
		{`{"jsonrpc":"2.0","id":{},"method":"ping"}`, `"code":-32600`},
		{`{`, `"code":-32600`},
	} {
		if w := invoke(s, tc.body, "Bearer "+token); !strings.Contains(w.Body.String(), tc.want) {
			t.Errorf("%s: %s", tc.body, w.Body.String())
		}
	}
	w := invoke(s, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, "Bearer "+token)
	if w.Code != http.StatusAccepted {
		t.Fatal(w.Code)
	}
}

func TestValidation(t *testing.T) {
	call := func(context.Context, string, string, json.RawMessage) (json.RawMessage, error) { return nil, nil }
	for _, tools := range [][]Tool{
		{{Name: "echo", InputSchema: json.RawMessage(`null`)}},
		{{Name: "echo", InputSchema: json.RawMessage(`{"type":"array"}`)}},
		{{Name: "echo", InputSchema: json.RawMessage(`{"type":"object"}`)}, {Name: "echo", InputSchema: json.RawMessage(`{"type":"object"}`)}},
	} {
		if _, err := New("frame", token, tools, call); err == nil {
			t.Fatal("invalid catalog accepted")
		}
	}
}

func TestReadHasPlainSourceAndSeparateMetadata(t *testing.T) {
	source := "\tfunc X() { println(\"&\") }\n"
	value, _ := json.Marshal(map[string]any{"path": "a.go", "text": source, "sha256": "hash"})
	s, e := New("frame", token, []Tool{{Name: "read", InputSchema: json.RawMessage(`{"type":"object"}`)}}, func(context.Context, string, string, json.RawMessage) (json.RawMessage, error) { return value, nil })
	if e != nil {
		t.Fatal(e)
	}
	w := invoke(s, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"read","arguments":{"path":"a.go"}}}`, "Bearer "+token)
	var reply struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if e := json.Unmarshal(w.Body.Bytes(), &reply); e != nil {
		t.Fatal(e)
	}
	if len(reply.Result.Content) != 2 || reply.Result.Content[1].Text != source || strings.Contains(reply.Result.Content[0].Text, `"text"`) {
		t.Fatal(w.Body.String())
	}
}
