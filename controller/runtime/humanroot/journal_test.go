package humanroot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	core "skyra/controller/core/humanroot"
)

func config(mission string) core.Config {
	return core.Config{Mission: mission, Human: "operator", Entrypoints: []core.Capability{"think", "act"}, Ceiling: []core.Capability{"read:workspace", "write:scratch"}}
}
func open(t *testing.T, dir string) *Journal {
	t.Helper()
	j, err := Open(dir, config("mission"))
	if err != nil {
		t.Fatal(err)
	}
	return j
}
func TestJournalHumanThinkHumanActAndRestart(t *testing.T) {
	dir := t.TempDir()
	j := open(t, dir)
	authored, err := j.Author("request", json.RawMessage(`{"goal":"make a plan"}`))
	if err != nil {
		t.Fatal(err)
	}
	request := authored.Observations[0]
	think, err := j.Invoke("think-call", "think", []core.ID{request}, []core.Capability{"read:workspace"})
	if err != nil {
		t.Fatal(err)
	}
	// The runtime bridge verifies receipt identity and resolves the actual result;
	// neither a model nor a human payload chooses its origin.
	reader := func(call core.Invocation, receipt string) ([]json.RawMessage, error) {
		if receipt != "accepted-think-return" || call.ID != think.Invocation.ID || call.Target != "think" {
			return nil, fmt.Errorf("invalid runtime receipt")
		}
		return []json.RawMessage{json.RawMessage(`{"goal":"write the plan","tools":["write:scratch"]}`)}, nil
	}
	if _, err = j.Receive(think.Invocation.ID, "forged", reader); err == nil {
		t.Fatal("unverified child output accepted")
	}
	got, err := j.Receive(think.Invocation.ID, "accepted-think-return", reader)
	if err != nil {
		t.Fatal(err)
	}
	spec := got.Observations[0]
	j = open(t, dir) // recover the human's received spec before forwarding it
	act, err := j.Invoke("act-call", "act", []core.ID{spec}, []core.Capability{"write:scratch"})
	if err != nil {
		t.Fatal(err)
	}
	if act.Kind != core.Dispatch || len(act.Passed) != 1 || act.Passed[0].ID() != spec || act.Passed[0].Source().Target != "think" {
		t.Fatal("handoff changed observation identity or source")
	}
	// A dispatched effect is recoverable but must never be replayed automatically.
	restarted := open(t, dir)
	state, err := restarted.State()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PendingCalls()) != 1 || state.PendingCalls()[0].ID != act.Invocation.ID {
		t.Fatal("recovery lost the pending child")
	}
	duplicate, err := restarted.Invoke("act-call", "act", []core.ID{spec}, []core.Capability{"write:scratch"})
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.Kind != core.Duplicate || duplicate.Invocation != nil || duplicate.Call != act.Invocation.ID {
		t.Fatal("recovery dispatched an unknown outcome again")
	}
}
func TestJournalRejectsForeignRootInputsAndReconfiguration(t *testing.T) {
	dir := t.TempDir()
	j := open(t, dir)
	other, err := Open(t.TempDir(), config("other"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := other.Author("request", json.RawMessage(`"same text"`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = j.Invoke("foreign", "act", result.Observations, nil); err == nil {
		t.Fatal("foreign Mission observation passed")
	}
	changed := config("mission")
	changed.Ceiling = append(changed.Ceiling, "write:home")
	if _, err = Open(dir, changed); err == nil {
		t.Fatal("pinned root was widened")
	}
	if _, err = j.Receive("unknown", "receipt", nil); err == nil {
		t.Fatal("receive needs a runtime verifier")
	}
}
func TestJournalDetectsTailTamperingAndMissingRecords(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			dir := t.TempDir()
			j := open(t, dir)
			if _, err := j.Author("a", json.RawMessage(`"original"`)); err != nil {
				t.Fatal(err)
			}
			if _, err := j.Author("b", json.RawMessage(`"later"`)); err != nil {
				t.Fatal(err)
			}
			files, _ := filepath.Glob(filepath.Join(dir, "records", "*.json"))
			if missing {
				if err := os.Remove(files[0]); err != nil {
					t.Fatal(err)
				}
			} else {
				b, err := os.ReadFile(files[1])
				if err != nil {
					t.Fatal(err)
				}
				var record core.Record
				if err = json.Unmarshal(b, &record); err != nil {
					t.Fatal(err)
				}
				record.Action.Value = json.RawMessage(`"tampered"`)
				b, _ = json.Marshal(record)
				if err = os.WriteFile(files[1], b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := j.State(); err == nil {
				t.Fatal("tampered journal accepted")
			}
		})
	}
}
func TestJournalSerializesIndependentAdaptersAndDuplicateRequests(t *testing.T) {
	dir := t.TempDir()
	a := open(t, dir)
	b := open(t, dir)
	var wg sync.WaitGroup
	for n := 0; n < 24; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			j := a
			if n%2 != 0 {
				j = b
			}
			if _, err := j.Author(fmt.Sprintf("request-%d", n%12), json.RawMessage(`"observation"`)); err != nil {
				t.Error(err)
			}
		}(n)
	}
	wg.Wait()
	state, err := a.State()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Scope()) != 12 || state.Sequence() != 12 {
		t.Fatalf("lost or repeated authoring: %d %d", len(state.Scope()), state.Sequence())
	}
}
