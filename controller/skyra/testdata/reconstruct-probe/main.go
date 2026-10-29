// An end-to-end fixture: the enclosing actor retains its own checkpoint,
// and later asks the real apprentice to compose a read from that checkpoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/targets/user/checkpoint"
	"skyra/controller/targets/user/projection"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	var e struct {
		Frame    string
		Contract json.RawMessage
		Input    map[string]json.RawMessage
		Steps    []struct {
			To, State, Observation string
			Value                  json.RawMessage
		}
	}
	if err := json.NewDecoder(os.Stdin).Decode(&e); err != nil {
		return err
	}
	emit := func(to string, payload any) {
		b, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
		fmt.Println(string(b))
	}
	store := checkpoint.Store{Dir: filepath.Join(os.Getenv("SKYRA_STATE"), "checkpoints")}
	if args, ok := e.Input["read-checkpoint"]; ok {
		var ask struct{ Conversation string }
		json.Unmarshal(args, &ask)
		cp, err := store.Read(ask.Conversation)
		if err != nil {
			return err
		}
		emit(e.Frame, cp)
		return nil
	}
	if args, ok := e.Input["project-state"]; ok {
		var ask struct {
			Conversation string
			Availability projection.Availability
			Checkpoint   checkpoint.Checkpoint
			Trace        projection.Tree
		}
		if err := json.Unmarshal(args, &ask); err != nil {
			return err
		}
		view, err := projection.ConversationState(ask.Conversation, ask.Checkpoint.RootInvocation, ask.Trace, ask.Availability)
		if err != nil {
			return err
		}
		emit(e.Frame, view)
		return nil
	}
	if args, ok := e.Input["project-app-state"]; ok {
		var ask struct {
			UIIncarnation string
			Revision      int
			Views         []projection.View
		}
		json.Unmarshal(args, &ask)
		snap, err := projection.AppState(ask.UIIncarnation, ask.Revision, ask.Views)
		if err != nil {
			return err
		}
		emit(e.Frame, snap)
		return nil
	}
	switch len(e.Steps) {
	case 0:
		if _, err := store.Initialize("c-probe", e.Frame, checkpoint.Contract{Target: "user", Schema: e.Contract}); err != nil {
			return err
		}
		fmt.Println(`[{"to":"echo","payload":{"text":"first"}},{"to":"echo","payload":{"text":"second"}}]`)
	case 2:
		cp, err := store.Read("c-probe")
		if err != nil {
			return err
		}
		emit("system", map[string]any{"trace": map[string]any{"invocation": cp.RootInvocation, "observations": []string{e.Steps[0].Observation, e.Steps[1].Observation}, "contract": cp.Contract}})
	case 3:
		if e.Steps[2].State != "return" {
			return fmt.Errorf("reconstruction syscall failed")
		}
		cp, err := store.Read("c-probe")
		if err != nil {
			return err
		}
		var ev checkpoint.Evidence
		if err := json.Unmarshal(e.Steps[2].Value, &ev); err != nil {
			return err
		}
		if _, err := store.Advance("c-probe", cp, ev); err != nil {
			return err
		}
		emit("apprentice", map[string]any{"reconstruct-state": map[string]any{"uiIncarnation": "fixture", "revision": 1, "conversations": []any{map[string]any{"conversation": "c-probe", "availability": projection.Availability{}}}}})
	case 4:
		if e.Steps[3].State != "return" {
			return fmt.Errorf("reconstruction procedure failed")
		}
		emit(e.Frame, e.Steps[3].Value)
	default:
		return fmt.Errorf("unexpected fixture step count")
	}
	return nil
}
