package dna_test

import (
	"context"
	"testing"

	dna "skyra-v1/src/skyra/skyra-dna-v14"
)

func TestPublicAPICompatibility(t *testing.T) {
	count := 0
	ct, err := dna.NewCTU("root", dna.DriverRegistration{Driver: &dna.RootDriver{EchoTarget: dna.Target{Address: "echo"}}, Input: dna.ProbeInput{}}, &dna.ProbeAuthority{CheckMove: func(context.Context, dna.Move) error { return nil }}, func(r dna.RootResult) error {
		count++
		if r.Err != nil || r.Observation == nil {
			t.Errorf("bad result: %+v", r)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ct.RegisterDriver("echo", dna.DriverRegistration{Driver: &dna.EchoDriver{}, Input: dna.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	id, err := dna.NewInvocationID()
	if err != nil {
		t.Fatal(err)
	}
	if err := ct.Submit(context.Background(), dna.Move{Target: dna.Target{Address: "root"}, Op: dna.Invoke{InvocationID: id, Action: dna.Action{Value: "echo"}, Params: []dna.Observation{{}}}}); err != nil {
		t.Fatal(err)
	}
	c, ok := ct.Snapshot(id)
	if !ok || !c.Resumed || c.Result == nil || count != 1 {
		t.Fatal("public API loop failed")
	}
}
