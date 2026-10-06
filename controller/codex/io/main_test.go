package main

import (
	"testing"

	"skyra/controller/runtime/hold"
)

func TestRequestsAreCheckedBeforeTheHolderIsAsked(t *testing.T) {
	for name, q := range map[string]hold.Request{
		"no incarnation": {Op: "exchange", WaitMS: 1},
		"unknown op":     {Incarnation: "i", Op: "hello", WaitMS: 1},
		"no wait":        {Incarnation: "i", Op: "exchange"},
		"too long":       {Incarnation: "i", Op: "exchange", WaitMS: 30001},
	} {
		if _, err := perform(q, "/private/tmp/none.sock"); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if _, err := perform(hold.Request{Incarnation: "i", Op: "exchange", WaitMS: 1}, ""); err == nil {
		t.Fatal("an operation with no granted resident was accepted")
	}
}

func TestAnUnreachableResidentIsGone(t *testing.T) {
	o, err := perform(hold.Request{Incarnation: "i", Op: "exchange", Deliver: []string{`{}`}, WaitMS: 10}, "/private/tmp/skyra-no-such-edge.sock")
	if err != nil || !o.PeerEOF || o.Delivery != "not_sent" || o.Lines == nil {
		t.Fatalf("observation: %+v %v", o, err)
	}
}
