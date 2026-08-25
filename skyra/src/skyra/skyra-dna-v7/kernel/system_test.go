package kernel_test

import (
	"path/filepath"
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	macos "skyra-v1/src/skyra/skyra-dna-v7/drivers/mac-os"
	"skyra-v1/src/skyra/skyra-dna-v7/kernel"
)

func TestActivateDefineAndRootOnlyRemember(t *testing.T) {
	rootName := c.Name{Value: "root"}
	childName := c.Name{Value: "child"}
	path := filepath.Join(t.TempDir(), "graph.json")
	driver := macos.New(path)
	if got := driver.Remember(c.Cone{Engrams: []c.Engram{{Name: rootName}, {Name: childName}}}); got != c.Success {
		t.Fatalf("seed: got %v", got)
	}
	system := kernel.New(driver, map[string]c.Account{"pin": {Root: c.Name{Value: "account"}}})

	account, horizon, outcome := system.Define(c.Pin{Secret: "pin"})
	if outcome != c.Success || account.Root != horizon.Root {
		t.Fatalf("define: account=%v horizon=%v outcome=%v", account, horizon, outcome)
	}
	root, outcome := system.Activate(rootName, horizon)
	if outcome != c.Success || root.Parent != nil {
		t.Fatalf("activate root: root=%v outcome=%v", root, outcome)
	}
	keep := c.Observation{Provenance: c.Provenance{Resource: childName, Invocation: "child-1"}}
	cone, outcome := system.Remember(root, keep)
	if outcome != c.Success || len(cone.Engrams) != 2 {
		t.Fatalf("remember cone: cone=%v outcome=%v", cone, outcome)
	}

	nonRoot := root
	nonRoot.Parent = &root
	if _, got := system.Remember(nonRoot, keep); got != c.Failure {
		t.Fatalf("non-root REMEMBER: got %v", got)
	}
	if _, got := macos.New(path).Activate(childName); got != c.Success {
		t.Fatalf("reopen durable graph: got %v", got)
	}
}
