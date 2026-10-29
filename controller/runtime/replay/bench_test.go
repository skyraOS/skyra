package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"skyra/controller/core/boundary"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// longLineage commits a version-2 run the way the runtime would: operator →
// agent, then rounds of agent → echo → echo's reply → resolve agent (each
// resolve carrying its round's complete emission), then the agent's reply and
// the resolve of the operator. It returns the head and the record count,
// 3*rounds + 4.
func longLineage(tb testing.TB, root string, rounds int) (string, int) {
	tb.Helper()
	for _, d := range []string{"records", "inputs"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			tb.Fatal(err)
		}
	}
	reach := map[string]boundary.Surface{"echo": echoS, "agent": agentS}
	g0, err := tree.Commit(root, record.Record{Version: 2, State: record.Invoke, Reach: reach})
	if err != nil {
		tb.Fatal(err)
	}
	h := boundary.Boundary{Stamp: g0, Reach: reach, Frame: g0}

	put := func(in []byte, r record.Record) string {
		if _, err := tree.Put(root, "inputs", in); err != nil {
			tb.Fatal(err)
		}
		s, err := tree.Commit(root, r)
		if err != nil {
			tb.Fatal(err)
		}
		return s
	}
	emit := func(under boundary.Boundary, out string) (string, boundary.Boundary) {
		_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldBoundary(under), In: []byte(out)})
		r := machine.CommittedRecord(effs)
		s := put([]byte(out), r)
		next, err := machine.Narrow(under, r, s)
		if err != nil {
			tb.Fatal(err)
		}
		return s, next
	}
	resolve := func(discharged boundary.Boundary, discharge, emission string) (string, boundary.Boundary) {
		w := discharged.Resume
		obs, _ := json.Marshal(machine.Observation{Frame: w.Frame, Round: w.Stamp, Discharges: []string{discharge}, Emission: emission})
		r, err := machine.Resolved(discharge, obs)
		if err != nil {
			tb.Fatal(err)
		}
		s := put(obs, r)
		resumed := *w
		resumed.Stamp = s
		return s, resumed
	}

	opCall := `{"to":"agent","payload":{"text":"go"}}`
	_, agent := emit(h, opCall)
	for i := 0; i < rounds; i++ {
		call := fmt.Sprintf(`{"to":"echo","payload":{"text":"step %d"}}`, i)
		_, echo := emit(agent, call)
		reply, done := emit(echo, fmt.Sprintf(`step %d`, i))
		_, agent = resolve(done, reply, record.Sum([]byte(call)))
	}
	reply, done := emit(agent, `done`)
	head, _ := resolve(done, reply, record.Sum([]byte(opCall)))
	return head, 3*rounds + 4
}

// inMemory loads a tree into a memory store, as the website hands a run to
// the browser: replay then measures derivation and hashing, not disk.
func inMemory(tb testing.TB, root string) memory {
	m := memory{}
	for _, kind := range []string{"records", "inputs"} {
		entries, err := os.ReadDir(filepath.Join(root, kind))
		if err != nil {
			tb.Fatal(err)
		}
		for _, e := range entries {
			b, err := os.ReadFile(filepath.Join(root, kind, e.Name()))
			if err != nil {
				tb.Fatal(err)
			}
			m[kind+"/"+e.Name()] = b
		}
	}
	return m
}

func TestLongLineageReplays(t *testing.T) {
	root := t.TempDir()
	head, want := longLineage(t, root, 50)
	if n, err := Replay(root, head); err != nil || n != want {
		t.Fatalf("replay = %d, %v; want %d", n, err, want)
	}
}

var benchRounds = []int{33, 333, 3333} // 103, 1,003 and 10,003 records

// BenchmarkReplayMemory replays a whole run from a store in memory.
func BenchmarkReplayMemory(b *testing.B) {
	for _, rounds := range benchRounds {
		root := b.TempDir()
		head, want := longLineage(b, root, rounds)
		m := inMemory(b, root)
		b.Run(fmt.Sprintf("records=%d", want), func(b *testing.B) {
			for b.Loop() {
				if n, err := Lineage(m, head); err != nil || n != want {
					b.Fatalf("lineage = %d, %v; want %d", n, err, want)
				}
			}
			b.ReportMetric(float64(want)*float64(b.N)/b.Elapsed().Seconds(), "records/s")
		})
	}
}

// BenchmarkReplayDisk replays a whole run from the tree on disk.
func BenchmarkReplayDisk(b *testing.B) {
	for _, rounds := range benchRounds {
		root := b.TempDir()
		head, want := longLineage(b, root, rounds)
		b.Run(fmt.Sprintf("records=%d", want), func(b *testing.B) {
			for b.Loop() {
				if n, err := Replay(root, head); err != nil || n != want {
					b.Fatalf("replay = %d, %v; want %d", n, err, want)
				}
			}
			b.ReportMetric(float64(want)*float64(b.N)/b.Elapsed().Seconds(), "records/s")
		})
	}
}
