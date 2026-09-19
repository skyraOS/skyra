package engine

import (
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v12/causal"
)

// Boot must retain the same resident frame for both provisioning and execution,
// including already-landed child observations and their declared order.
func TestBootSharesPreloadedMemory(t *testing.T) {
	// Opaque identity fixtures: canonical encoding is not implemented by this test.
	parent := MoveID{Value: "parent"}
	first := MoveID{Value: "original"}
	second := MoveID{Value: "revised"}
	observation := &c.Observation{Content: c.Content{Bytes: []byte("revised document")}}
	frame := &Register{Children: []ChildObservation{
		{Fd: first},
		{Fd: second, Value: observation},
	}}
	memory := &Memory{Fds: map[MoveID]*Fd{parent: {Move: Move{Fd: parent}, Register: frame}}}
	runtime := &Runtime{Memory: memory, Kernel: &Kernel{}, CTU: &CTU{}}
	runtime.Boot()

	if runtime.Kernel.Memory != memory || runtime.CTU.Memory != memory {
		t.Fatal("kernel provisioning and CTU execution must use the same Memory instance")
	}
	got := runtime.CTU.Memory.Fds[parent].Register
	if got != frame || runtime.Kernel.Memory.Fds[parent].Register != frame ||
		runtime.CTU.Memory.Fds[parent].Move.Fd != parent {
		t.Fatal("boot replaced a preloaded register")
	}
	if len(got.Children) != 2 || got.Children[0].Fd != first || got.Children[0].Value != nil ||
		got.Children[1].Fd != second || got.Children[1].Value != observation {
		t.Fatal("boot changed declared child order or existing observation state")
	}
}
