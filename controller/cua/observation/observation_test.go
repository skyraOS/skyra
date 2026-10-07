package observation

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRequestCannotExpressInputOrEscapeBounds(t *testing.T) {
	for _, b := range []string{
		`null`, `{"pid":1}`, `{"pid":1.5,"window_id":2}`, `{"pid":1,"window_id":2,"tool":"click"}`,
		`{"pid":1,"window_id":2,"screenshot_out_file":"/tmp/x"}`, `{"pid":1,"window_id":2,"include_screenshot":"true"}`,
		`{"pid":1,"window_id":2,"max_elements":251}`, `{"pid":1,"window_id":2,"max_depth":26}`,
		`{"pid":1,"window_id":2} {}`, `{"pid":1,"window_id":2,"max_elements":-1}`,
	} {
		if _, err := DecodeRequest([]byte(b)); err == nil {
			t.Errorf("accepted %s", b)
		}
	}
	if _, err := DecodeRequest([]byte(`{"pid":1,"window_id":2,"query":"Count"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestProjectionPreservesRefusalAndObservationStatus(t *testing.T) {
	scope := Scope{PID: 7, WindowID: 9}
	r := Project(scope, []byte(`{"content":[{"type":"text","text":"Count: 42"}],"structuredContent":{"snapshot_id":"snap","truncated":true,"degraded":true}}`))
	if r.Error != nil || r.Text != "Count: 42" || r.SnapshotID != "snap" || !r.Truncated || !r.Degraded {
		t.Fatalf("%+v", r)
	}
	r = Project(scope, []byte(`{"isError":true,"content":[{"type":"text","text":"refused"}],"structuredContent":{"refusal":{"code":"window_owner_pid_mismatch","message":"wrong owner"}}}`))
	if r.Error == nil || r.Error.Code != "window_owner_pid_mismatch" || r.Error.Message != "wrong owner" {
		t.Fatalf("%+v", r)
	}
	for _, b := range []string{`{}`, `{"content":[{"type":"image","data":"secret"}]}`, `not json`} {
		if r := Project(scope, []byte(b)); r.Error == nil {
			t.Fatalf("accepted %s", b)
		}
	}
}

func TestObservationTextIsBoundedAtUTF8Boundary(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"content": []any{map[string]string{"type": "text", "text": strings.Repeat("é", MaxText)}}})
	r := Project(Scope{}, b)
	if r.Error != nil || !r.Truncated || len(r.Text) != MaxText {
		t.Fatalf("%+v", r)
	}
	if !json.Valid(mustJSON(t, r)) {
		t.Fatal("truncated text broke JSON")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFramesRequireNewlineAndRespectLimit(t *testing.T) {
	for _, b := range []string{"{}", strings.Repeat("x", 10) + "\n"} {
		if _, err := ReadLine(strings.NewReader(b), 10); err == nil {
			t.Fatalf("accepted %q", b)
		}
	}
	if b, err := ReadLine(bytes.NewBufferString("{}\n"), 10); err != nil || string(b) != "{}\n" {
		t.Fatalf("%q: %v", b, err)
	}
}

func TestManifestHasOnlyWindowObservation(t *testing.T) {
	b, err := Manifest("com.example.Fixture")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Version int `json:"version"`
		Allow   struct {
			Tools []string `json:"tools"`
		} `json:"allow"`
		Resources struct {
			Desktop struct {
				Display bool `json:"display"`
			} `json:"desktop"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.Version != 3 || len(m.Allow.Tools) != 1 || m.Allow.Tools[0] != "get_window_state" || m.Resources.Desktop.Display {
		t.Fatalf("%s", b)
	}
	if ManifestHash(b) == ManifestHash(append(b, '\n')) {
		t.Fatal("hash failed to bind exact bytes")
	}
}
