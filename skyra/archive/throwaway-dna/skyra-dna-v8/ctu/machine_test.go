package ctu_test

import (
	"bytes"
	"path/filepath"
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
	"skyra-v1/src/skyra/skyra-dna-v8/ctu"
	"skyra-v1/src/skyra/skyra-dna-v8/drivers"
	"skyra-v1/src/skyra/skyra-dna-v8/entity"
	"skyra-v1/src/skyra/skyra-dna-v8/fabric"
	"skyra-v1/src/skyra/skyra-dna-v8/kernel"
)

func TestPossibilityPartitionsAcrossTwoCTUsAndRememberGathersTheCone(t *testing.T) {
	clientName := c.Name{Value: "client"}
	workerName := c.Name{Value: "worker"}
	seedName := c.Name{Value: "seed"}
	echoName := c.Name{Value: "echo"}
	accountName := c.Name{Value: "account"}

	security, outcome := fabric.NewSecurity(clientName, workerName)
	if outcome != c.Success {
		t.Fatal("create security")
	}
	authority := security.Authority()
	verifier := security.Verifier()
	clientCustody, outcome := security.Custody(clientName)
	if outcome != c.Success {
		t.Fatal("client custody")
	}
	workerCustody, outcome := security.Custody(workerName)
	if outcome != c.Success {
		t.Fatal("worker custody")
	}
	transport := fabric.NewLocal(authority, verifier)
	driver := drivers.NewFile(filepath.Join(t.TempDir(), "graph.json"))
	if got := driver.Seed(
		c.Engram{Name: seedName},
		c.Engram{Name: echoName, Contracts: []c.Contract{{Name: echoName}}},
	); got != c.Success {
		t.Fatalf("seed durable graph: %v", got)
	}
	system := kernel.New(driver, transport, verifier, authority, security.Minter(), map[string]c.Account{
		"pin": {Root: accountName},
	})
	_, horizon, outcome := system.Define(c.Pin{Secret: "pin"})
	if outcome != c.Success || authority.VerifyHorizon(horizon) != c.Success {
		t.Fatalf("mint root horizon: %v", outcome)
	}

	client := ctu.New("client", transport, clientCustody, authority, system)
	worker := ctu.New("worker", transport, workerCustody, authority, nil)
	if transport.Register("client", client) != c.Success || transport.Register("worker", worker) != c.Success {
		t.Fatal("register CTUs")
	}
	if client.RegisterEntity(seedName, c.Name{Value: "seed-unit"}, entity.Function(
		func(_ []c.Observation, payload c.Payload) (c.Content, c.Outcome) {
			return c.Content{Bytes: append([]byte("local:"), payload.Body...)}, c.Success
		},
	)) != c.Success {
		t.Fatal("register local entity")
	}
	if worker.RegisterEntity(echoName, c.Name{Value: "echo-unit"}, entity.Function(
		func(operands []c.Observation, payload c.Payload) (c.Content, c.Outcome) {
			body := []byte("remote:")
			if len(operands) > 0 {
				body = append(body, operands[0].Content.Bytes...)
				body = append(body, ':')
			}
			return c.Content{Bytes: append(body, payload.Body...)}, c.Success
		},
	)) != c.Success || client.RegisterRemote(echoName, c.Target{Endpoint: "worker"}) != c.Success {
		t.Fatal("register remote entity")
	}

	seed := &c.Possibility{
		ID: c.PossibilityID{Value: "seed-fragment"}, Arena: c.Arena{Root: accountName},
		Name: seedName, Horizon: horizon, Payload: c.Payload{Body: []byte("cause")},
	}
	if got := client.Execute(seed); got != c.Success || seed.Result == nil {
		t.Fatalf("execute local cause: %v", got)
	}

	root, outcome := system.Activate(echoName, horizon)
	if outcome != c.Success {
		t.Fatalf("activate root: %v", outcome)
	}
	root.Payload = c.Payload{Body: []byte("work")}
	root.Frame.Scope = append(root.Frame.Scope, *seed.Result)
	if got := client.Execute(&root); got != c.Success {
		t.Fatalf("execute distributed possibility: %v", got)
	}
	if root.Result == nil || !bytes.Equal(root.Result.Content.Bytes, []byte("remote:local:cause:work")) {
		t.Fatalf("returned observation: %+v", root.Result)
	}
	if root.Result.Segment.By != workerName || verifier.Verify(root.Result.Segment) != c.Success {
		t.Fatalf("remote segment was not signed by worker: %+v", root.Result.Segment)
	}
	if len(root.Result.Segment.Provenance.Causes) != 1 ||
		root.Result.Segment.Provenance.Causes[0] != seed.Result.Segment.Ref {
		t.Fatalf("cross-cut lineage: %+v", root.Result.Segment.Provenance.Causes)
	}
	if got := len(client.Orient(&root).InFlight); got != 0 {
		t.Fatalf("frontier did not fold at RECOGNIZE: %d", got)
	}

	cone, outcome := system.Remember(&root, *root.Result)
	if outcome != c.Success || len(cone.Segments) != 2 {
		t.Fatalf("gathered cone: outcome=%v cone=%+v", outcome, cone)
	}
	cut := c.MoveID{Value: "not-a-root"}
	nonRoot := root
	nonRoot.Parent = &cut
	if _, got := system.Remember(&nonRoot, *root.Result); got != c.Failure {
		t.Fatalf("remote fragment acquired commit authority: %v", got)
	}
}

func TestSignedHorizonAndProvenanceRejectTampering(t *testing.T) {
	ctuName := c.Name{Value: "ctu"}
	security, outcome := fabric.NewSecurity(ctuName)
	if outcome != c.Success {
		t.Fatal("create security")
	}
	authority := security.Authority()
	horizon, outcome := security.Minter().Mint(c.Name{Value: "account"})
	if outcome != c.Success {
		t.Fatal("mint")
	}
	narrowed, outcome := authority.Attenuate(horizon, c.Caveat{Constraint: "slice:edge"})
	if outcome != c.Success || authority.VerifyHorizon(narrowed) != c.Success {
		t.Fatalf("attenuate: %v", outcome)
	}
	tamperedHorizon := narrowed
	tamperedHorizon.Caveats = append([]c.Caveat(nil), narrowed.Caveats...)
	tamperedHorizon.Caveats[0].Constraint = "widened"
	if authority.VerifyHorizon(tamperedHorizon) != c.Failure {
		t.Fatal("forged caveat admitted")
	}

	custody, outcome := security.Custody(ctuName)
	if outcome != c.Success {
		t.Fatal("custody")
	}
	segment, outcome := custody.Sign(c.Segment{
		By: ctuName, Provenance: c.Provenance{Resource: c.Name{Value: "echo"}, Invocation: "edge"},
	})
	if outcome != c.Success || security.Verifier().Verify(segment) != c.Success {
		t.Fatalf("sign segment: %v", outcome)
	}
	tamperedSegment := segment
	tamperedSegment.Provenance.Invocation = "other-edge"
	if security.Verifier().Verify(tamperedSegment) != c.Failure {
		t.Fatal("forged provenance admitted")
	}
}
