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
// D1 · a naming commit carries ALL the reach it was handed
// ---------------------------------------------------------------------------

func TestNamingCommitCarriesEveryReach(t *testing.T) {
	// A resolver hands back two names. Its parent is an orchestrator, so it can
	// only NAME what came back — and the naming commit is where reach used to be
	// silently narrowed to one, because every EngramRef was written to the same
	// content key.
	//
	// Delegation is the only way these two names can be reached: they are minted
	// into the index but are NOT builtin, so nothing puts them in a frame except
	// a child handing them back.
	g := graph.New()
	reg := entity.NewRegistry()

	var resolved []string

	table := scripted.Table{
		"top": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "middle", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			// The frame AFTER the group closed. Both delegated names must be
			// nameable here, or a capability was lost in transit.
			for _, n := range []string{"alpha", "beta"} {
				if _, ok := f.Resolve(n); ok {
					resolved = append(resolved, n)
				}
			}
			return commitNaming(f)
		},
		// An ORCHESTRATOR: it selects, so it may only name what came back. The
		// reach it passes up is the reach its child returned.
		"middle": func(f frame.Frame, step int) (grammar.Emission, error) {
			if step == 0 {
				in := f.Observations()[0]
				return grammar.Select{Members: []grammar.Member{{
					Engram: "resolver", Cite: in.ID, Entity: ent,
					Observations: []ids.ObservationID{in.ID},
				}}}, nil
			}
			return grammar.Commit{Observed: f.Returned()[0].ID}, nil
		},
	}

	reg.Register(ent, scripted.New(table))
	k := kernel.New(g, reg, &host.Local{Root: t.TempDir()}, nil)

	// The two capabilities. Minted so the kernel can resolve them, and NOT
	// builtin, so delegation is the only path by which they arrive.
	alpha := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "do the first thing"})
	beta := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "do the second thing"})
	g.Mint(alpha)
	g.Mint(beta)

	table["resolves"] = leaf(content.Content{
		"found": content.Bool(true),
		"reach": content.List{
			content.EngramRef{ID: alpha.ID, Name: "alpha"},
			content.EngramRef{ID: beta.ID, Name: "beta"},
		},
	})
	resolver := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "resolves"})
	middle := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: "middle"})
	g.Mint(resolver)
	g.Mint(middle)
	k.Builtin["resolver"] = resolver
	k.Builtin["middle"] = middle

	obs := k.Ingress(content.Content{"need": content.Str("two capabilities")})
	if _, _, err := k.Ask("top", ent, []*graph.Observation{obs}); err != nil {
		t.Fatalf("ask: %v", err)
	}

	if len(resolved) != 2 {
		t.Fatalf("the parent could name %v, want both alpha and beta — "+
			"a naming commit dropped delegated reach silently", resolved)
	}
	t.Logf("both delegated names resolved: %v", resolved)
}

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
