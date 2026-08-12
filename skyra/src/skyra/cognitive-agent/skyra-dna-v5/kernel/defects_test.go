package kernel_test

// Regression tests for defects found by reading the code rather than by running
// it. They are kept apart from substrate_test.go on purpose: that file asserts
// the design's CLAIMS, and these assert that two specific bugs stay fixed.

import (
	"fmt"
	"strings"
	"testing"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity/scripted"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/host"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/kernel"
)

// ---------------------------------------------------------------------------
// D2 · a write the driver refuses is still DISPOSED
// ---------------------------------------------------------------------------

// faultyHost implements the storage class and refuses every write. The point is
// that an absent implementation and a failing one are different: this host says
// yes to the class and no to the call, which is the case Materialize dropped.
type faultyHost struct{ err error }

func (h faultyHost) ID() ids.HostID               { return "faulty" }
func (h faultyHost) Execution() host.ExecutionOps { return nil }
func (h faultyHost) Storage() host.StorageOps     { return faultyStore{h.err} }

type faultyStore struct{ err error }

func (s faultyStore) Read(path string) ([]byte, error)     { return nil, s.err }
func (s faultyStore) Write(path string, body []byte) error { return s.err }

func TestFailedWriteIsDisposedAndRecorded(t *testing.T) {
	// The rule is that EVERY staged effect is explicitly disposed, and that an
	// effect which did not happen is a reportable fact rather than a no-op.
	// Unreachable was handled; REACHED-AND-REFUSED was not, and fell into a
	// fourth state nothing could name and nothing recorded.
	g := graph.New()
	reg := entity.NewRegistry()
	h := faultyHost{err: fmt.Errorf("disk is read-only")}

	var root *kernel.Root

	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "writer", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			// Cite it, so the credit walk REACHES the staging engram. The
			// effect is not unreachable; the driver is what refuses it.
			return grammar.Commit{Observed: f.Returned()[0].ID}, nil
		},
		"writes": func(f frame.Frame, step int) (grammar.Emission, error) {
			handle := root.Stage(&kernel.Effect{
				Engram: f.Self(), Class: "storage",
				Path: "wanted.txt", Body: []byte("this was asked for"),
			})
			return grammar.Commit{Content: content.Content{
				"wrote": content.WorldRef{Host: "faulty", Handle: handle},
			}}, nil
		},
	}

	reg.Register(ent, scripted.New(table))
	k := kernel.New(g, reg, h, nil)
	writer := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "writes"})
	g.Mint(writer)
	k.Builtin["writer"] = writer

	obs := k.Ingress(content.Content{"need": content.Str("write a file")})

	rootEngram := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "top"})
	root = &kernel.Root{Span: "span-d2", Engram: rootEngram}

	out, err := k.Drive(root, ent, []*graph.Observation{obs})
	if err != nil {
		t.Fatalf("drive: %v", err)
	}

	// Accept surfaces the driver's error rather than swallowing it.
	if err := k.Accept(root, out); err == nil {
		t.Fatal("the driver refused the write and Accept reported success")
	}

	staged := root.Staged()
	if len(staged) != 1 {
		t.Fatalf("want 1 staged effect, got %d", len(staged))
	}
	if d := staged[0].Disposition; d != kernel.Failed {
		t.Errorf("effect ended as %v, want failed — a staged effect must never "+
			"end in a state nothing can name", d)
	}

	ds := g.Discards()
	if len(ds) != 1 {
		t.Fatalf("want 1 recorded discard, got %d — a lost write must never be a silent no-op", len(ds))
	}
	if !strings.Contains(ds[0].Detail, "read-only") {
		t.Errorf("the discard does not carry the driver's reason: %q", ds[0].Detail)
	}
	if ds[0].Engram != writer.ID {
		t.Errorf("the discard names %s, want the engram that staged it (%s)", ds[0].Engram, writer.ID)
	}
	t.Logf("recorded: %s", ds[0].Detail)
}
